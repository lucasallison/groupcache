package groupcache

import (
	"bytes"
	"context"
	"encoding/gob"
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

// TODO add more fields? See the response struct
type CachedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

type ProxyCache struct {
	etagger *tagger.ETagger
	group   *Group
}

// TODO name? now only pc ...
func NewProxyCache(cacheBytes int64) *ProxyCache {
	pc := ProxyCache{
		etagger: tagger.NewTagger(nil, false),
		group: NewGroup("pc", cacheBytes, GetterFunc(
			func(ctx Context, key string, dest Sink) error {
				return nil
			})),
	}
	return &pc
}

func (pc *ProxyCache) Get(ctx context.Context, proxy ProxyWrapper) error {

	key := proxy.Req.URL.Path

	proxy.Proxy.ModifyResponse = func(r *http.Response) error {

		modified := true
		cachehit := false
		cachedBytes := []byte{}
		dest := AllocatingByteSliceSink(&cachedBytes)

		/* NOT MODIFIED: try to load from cache */
		if r.StatusCode == http.StatusNotModified {

			modified = false
			value, cachehit := pc.group.lookupCache(key)

			/* CACHE HIT: try to load from cache */
			if cachehit {
				logInfo(cachehit, modified, key)
				return pc.processCacheHit(r, &cachedBytes, dest, &value)
			}

		}

		logInfo(cachehit, modified, key)

		/* OBJECT MODIFIED OR CACHE MISS: update cache */
		return pc.modifyCache(dest, key, r)
	}

	pc.serveRequest(key, proxy)

	return nil
}

func (pc *ProxyCache) serveRequest(key string, proxy ProxyWrapper) {

	proxy.Req.Header.Set("ETag", pc.etagger.GetEtag(key))

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

	// TODO see line 298 of groupecache. We cant only call this function...
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

func replaceResponse(r *http.Response, cr *CachedResponse) error {
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

func decodeResponse(b []byte) (*CachedResponse, error) {
	/* decode cached response */
	cr := CachedResponse{}
	dec := gob.NewDecoder(bytes.NewReader(b))
	err := dec.Decode(&cr)
	return &cr, err
}

func encodeResponse(r *http.Response) (*[]byte, error) {

	respbody, err := getBodyAsBytes(r)

	/* encode struct */
	cr := CachedResponse{
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

func getBodyAsBytes(r *http.Response) (*[]byte, error) {

	/* read body */
	defer r.Body.Close()
	respbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	/* replace the cleared out body */
	r.Body = ioutil.NopCloser(bytes.NewBuffer(respbody))

	return &respbody, nil
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
