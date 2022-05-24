package groupcache

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"

	tagger "github.com/lucasallison/ETagger"
)

type ProxyWrapper struct {
	Proxy  *httputil.ReverseProxy
	Writer http.ResponseWriter
	Req    *http.Request
}

type cachedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

type ProxyCache struct {
	etagger   *tagger.ETagger
	group     *Group
	forwarder *forwarder
	validate  bool
}

var invalidationPools = [][]string{
	{"user"},
	{"category", "path"},
	{"path", "category", "skill"},
	{"skill", "path", "resource"},
	{"resource", "skill"},
	{"blog"},
}

func NewProxyCache(cacheBytes int64, validate bool, ctype string) *ProxyCache {
	pc := ProxyCache{
		etagger: tagger.NewTagger(invalidationPools),
		group: NewGroup("pc", cacheBytes, GetterFunc(
			func(ctx Context, key string, dest Sink) error {
				return nil
			})),
		forwarder: newForwarder(),
		validate:  validate,
	}

	// Set the cache type
	pc.group.mainCache.ctype = ctype
	pc.group.hotCache.ctype = ctype

	return &pc
}

func (pc *ProxyCache) RegisterPeerGroup(self string, peers ...string) {
	pc.forwarder.set(self, peers...)
}

func (pc *ProxyCache) Get(ctx context.Context, proxy ProxyWrapper) error {

	key := proxy.Req.URL.Path
	_, cachehit := pc.group.lookupCache(key)

	/* if it is not stored locally, check if it is stored in a peer */
	if !cachehit {
		res, err, ok := pc.forwarder.Forward(key, proxy.Req)
		if err != nil {
			proxy.Writer.WriteHeader(http.StatusInternalServerError)
			return err
		}

		if ok {
			writeRecievedResponse(proxy.Writer, res)
			return nil
		}
	}

	v := validator{}
	if pc.validate {
		v.makeRequest(proxy.Req)
	}

	proxy.Proxy.ModifyResponse = func(r *http.Response) error {

		if r.StatusCode >= http.StatusBadRequest {
			return nil
		}

		modified, cachehit, cachedBytes := true, false, []byte{}
		dest := AllocatingByteSliceSink(&cachedBytes)

		/* NOT MODIFIED: load from cache */
		if r.StatusCode == http.StatusNotModified {

			modified = false
			cachehit = true
			value, _ := pc.group.lookupCache(key)

			logInfo(cachehit, modified, key)
			err := pc.processCacheHit(r, &cachedBytes, dest, &value)

			if pc.validate {
				v.validate(r)
			}
			return err
		}

		/* OBJECT MODIFIED OR CACHE MISS: update cache */
		logInfo(cachehit, modified, key)
		return pc.modifyCache(dest, key, r)
	}

	pc.serveRequest(key, proxy, cachehit)
	return nil
}

func (pc *ProxyCache) serveRequest(key string, proxy ProxyWrapper, storedInCache bool) {

	/* the etagger might contain entries that have been evicted */
	if storedInCache {
		proxy.Req.Header.Set("ETag", pc.etagger.GetEtag(key))
	}

	proxy.Proxy.ServeHTTP(proxy.Writer, proxy.Req)

	/* dont modify future results e.g. for post requests */
	proxy.Proxy.ModifyResponse = nil
}

/*
	modify cache saves the etag and encodes the reponse.
	writeToCache only writes the supplied bytes to the cache.
*/
func (pc *ProxyCache) modifyCache(dest Sink, key string, r *http.Response) error {

	// TODO should we base the etag on the entire request? not just the body?
	/* update ETag */
	pc.etagger.SaveResponseAsTag(key, r)

	encResp, err := encodeResponse(r)
	if err != nil {
		return err
	}

	return pc.writeToCache(key, dest, *encResp)
}

func (pc *ProxyCache) writeToCache(key string, dest Sink, b []byte) error {
	err := dest.SetBytes(b)
	if err != nil {
		return err
	}

	value, err := dest.view()
	if err != nil {
		return err
	}

	pc.group.populateCache(key, value, &pc.group.mainCache)

	return nil
}

func (pc *ProxyCache) processCacheHit(r *http.Response, cachedBytes *[]byte, dest Sink, value *ByteView) error {

	/* will write the bytes of value (cached response) into cachedBytes */
	err := setSinkView(dest, *value)
	if err != nil {
		replaceWithInternalServerError(r)
		return err
	}

	cr, err := decodeResponse(*cachedBytes)
	if err != nil {
		replaceWithInternalServerError(r)
		return err
	}

	return replaceResponse(r, cr)
}

func replaceWithInternalServerError(r *http.Response) {
	r.StatusCode = http.StatusInternalServerError
	// TODO make this better ...
	// r.Body = nil
}

func replaceResponse(r *http.Response, cr *cachedResponse) error {
	/* clear existing headers */
	r.Header = make(http.Header)

	/* set cached headers */
	for name, values := range cr.Header {
		for _, value := range values {
			r.Header.Set(name, value)
		}
	}

	/* replace existing body with cached body */
	r.Body = io.NopCloser(bytes.NewReader(cr.Body))

	/* replace existing status code with cached status code */
	r.StatusCode = cr.StatusCode

	/* status??? */
	// r.status = ...

	return nil
}

func decodeResponse(b []byte) (*cachedResponse, error) {
	/* decode cached response */
	cr := cachedResponse{}
	dec := gob.NewDecoder(bytes.NewReader(b))
	err := dec.Decode(&cr)
	return &cr, err
}

func encodeResponse(r *http.Response) (*[]byte, error) {

	respbody, err := getBodyAsBytes(&r.Body)

	/* encode struct */
	cr := cachedResponse{
		StatusCode: r.StatusCode,
		Header:     r.Header,
		Body:       *respbody,
	}

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(cr)
	if err != nil {
		return nil, err
	}

	encodedResponse := buf.Bytes()
	return &encodedResponse, nil
}

func getBodyAsBytes(body *io.ReadCloser) (*[]byte, error) {

	/* read body */
	defer (*body).Close()
	bodyAsBytes, err := ioutil.ReadAll(*body)
	if err != nil {
		return nil, err
	}
	/* replace the cleared out body */
	*body = ioutil.NopCloser(bytes.NewBuffer(bodyAsBytes))

	return &bodyAsBytes, nil
}

func writeRecievedResponse(w http.ResponseWriter, res *http.Response) {

	bodyAsBytes, err := getBodyAsBytes(&res.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(res.StatusCode)
	for h, val := range res.Header {
		for _, v := range val {
			fmt.Println(h, v)
			w.Header().Add(h, v)
		}
	}
	fmt.Println(w.Header())

	w.Write(*bodyAsBytes)
}

func logInfo(cachehit bool, modified bool, key string) {
	if cachehit {
		log.Println("CACHE HIT! For ", key)
	} else if !modified {
		log.Println("CACHE MISS! Updating cache for ", key)
	} else {
		log.Println("MODIFIED! Updating cache for ", key)
	}
}
