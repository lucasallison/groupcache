package groupcache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TODO naming ....

type HTTPFetcher struct {
	// Pass the context of an incoming request
	Context *gin.Context

	// Fetch results saved here
	// TODO set to nil and check this
	resp *http.Response

	// TODO should this be saved here?
	respbody *[]byte
}

// TODO comments
func (h *HTTPFetcher) fetch(key string) error {

	// TODO remove log
	log.Printf("CACHE MISS (key: %s)", key)

	var err error

	h.resp, err = http.Get(key)
	if err != nil {
		return err
	}

	return nil
}

func (h *HTTPFetcher) getRespBody(des Sink) error {
	// TODO implement?
	return nil
}

func (h *HTTPFetcher) readRespBody() error {

	defer h.resp.Body.Close()
	respbody, err := ioutil.ReadAll(h.resp.Body)
	if err != nil {
		return err
	}

	h.respbody = &respbody
	return nil
}

func (h *HTTPFetcher) prepareContextForRepsonse() {
	h.Context.Data(h.resp.StatusCode, h.resp.Header.Get("Content-Type"), *h.respbody)
}

func (h *HTTPFetcher) determineCachability(dest Sink) error {

	// Do not cache
	if h.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("object not cached: non 200 status code recieved")
	}

	dest.SetBytes(*h.respbody)
	return nil
}

func InitWebProxyGroup(cacheBytes int64) *Group {
	return NewGroup("wp", cacheBytes, GetterFunc(
		func(ctx Context, key string, dest Sink, hf HTTPFetcher) error {

			err := hf.fetch(key)
			if err != nil {
				return err
			}

			err = hf.readRespBody()
			if err != nil {
				return err
			}

			hf.prepareContextForRepsonse()

			err = hf.determineCachability(dest)
			if err != nil {
				return err
			}

			return nil
		},
	))
}
