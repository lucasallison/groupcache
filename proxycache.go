package groupcache

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)

// TODO naming ....

func responseModifier(r *http.Response) error {

	// Copy the stream
	var buf bytes.Buffer
	tee := io.TeeReader(r.Body, &buf)

	log.Println(ioutil.ReadAll(tee))
	log.Println(ioutil.ReadAll(&buf))

	/*
		defer r.Body.Close()
		respbody, err := ioutil.ReadAll(tee)
		if err != nil {
			return err
		}

		return nil
	*/
	fmt.Println("yes!")
	return nil
}

type ProxyFecher struct {
	proxy *httputil.ReverseProxy
	w     http.ResponseWriter
	r     *http.Request
}

func InitWebProxyGroup(cacheBytes int64) *Group {
	return NewGroup("wp", cacheBytes, GetterFunc(
		func(ctx Context, key string, dest Sink, pf ProxyFecher) error {

			pf.proxy.ModifyResponse = responseModifier
			pf.proxy.ServeHTTP(pf.w, pf.r)

			// body...
			var b []byte
			dest.SetBytes(b)
			return nil
		},
	))
}
