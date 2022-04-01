package groupcache

import (
	"net/http"
	"net/http/httputil"
)

type ProxyFetcher struct {
	Proxy  *httputil.ReverseProxy
	Writer http.ResponseWriter
	Req    *http.Request
}

func NewProxyCacheGroup(cacheBytes int64) *Group {
	ng := NewGroup("pc", cacheBytes, GetterFunc(
		func(ctx Context, key string, dest Sink) error {
			/* the validator acts as the getter */
			return nil
		}))
	ng.proxyCache = true
	ng.validator = NewValidator()
	return ng
}
