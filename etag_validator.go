package groupcache

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/gob"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type Validator struct {
	mu sync.RWMutex
	/* map from key to its hash */
	hashes map[string][]byte
}

func NewValidator() *Validator {
	v := Validator{}
	v.hashes = make(map[string][]byte)
	return &v
}

func (v *Validator) calcHash(b []byte) []byte {
	hasher := sha1.New()
	hasher.Write(b)
	return hasher.Sum(nil)
}

func (v *Validator) SaveEtag(URI string, value string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.hashes[URI] = v.calcHash([]byte(value))
}

func (v *Validator) getEtag(URI string) string {
	return string(v.hashes[URI])
}

/* check if the supplied hash mathes the hash associated with the key */
func (v *Validator) ValidateLocal(URI string, etag string) bool {
	return v.getEtag(URI) == etag
}

// TODO naming
func GetRemote(g *Group, ctx context.Context, key string, pf ProxyFetcher) error {
	pf.Proxy.ModifyResponse = func(r *http.Response) error {

		b := []byte{}
		dest := AllocatingByteSliceSink(&b)

		/* NOT MODIFIED: try to load from cache */
		if r.StatusCode == http.StatusNotModified {

			value, cacheHit := g.lookupCache(key)

			if cacheHit {
				/* CACHE HIT: try to load from cache */
				log.Println("CACHE HIT! For ", key)

				// will write the bytes into b
				err := setSinkView(dest, value)
				if err != nil {
					replaceWithInternalServerError(r)
					return err
				}

				cr, err := decodeResponse(b)
				if err != nil {
					replaceWithInternalServerError(r)
					return err
				}

				err = replaceResponse(r, cr)
				if err != nil {
					return err
				}
			}

			/* CACHE MISS: update cache with fetched data */
			log.Println("CACHE MISS! For ", key)

		}

		/* OBJECT MODIFIED: update cache */
		log.Println("Updating: ", key)

		encResp, err := encodeResponse(r)
		if err != nil {
			return err
		}

		return writeToCache(g, key, dest, *encResp)
	}

	pf.Proxy.ServeHTTP(pf.Writer, pf.Req)

	return nil
}

func replaceWithInternalServerError(r *http.Response) {
	r.StatusCode = http.StatusInternalServerError
	// TODO make this better ...
	r.Body = nil
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
	/* read body */
	defer r.Body.Close()
	respbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(respbody))

	/* encode */
	cr := CachedResponse{
		StatusCode: r.StatusCode,
		Header:     r.Header,
		Body:       respbody,
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

func writeToCache(g *Group, key string, dest Sink, b []byte) error {
	err := dest.SetBytes(b)
	if err != nil {
		return err
	}

	value, err := dest.view()
	if err != nil {
		return err
	}

	// TODO see line 298 of groupecache. We cant only call this function...
	g.populateCache(key, value, &g.mainCache)

	return nil
}

// TODO add more fields? See the response struct
type CachedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}
