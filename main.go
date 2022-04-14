package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	groupcache "github.com/golang/groupcache/cache"
)

//var cacheBytes int64 = 64 << 20
var cacheBytes int64 = 1500
var proxyCache = groupcache.NewProxyCache(cacheBytes)

var proxy = httputil.NewSingleHostReverseProxy(&url.URL{})

func director(r *http.Request) {
	r.Host = GetHostFromEnv()
	r.URL.Host = r.Host
	r.URL.Scheme = "http"
}

func serveRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		/* Check the proxy cache for GET requests */
		pw := groupcache.ProxyWrapper{Proxy: proxy, Writer: w, Req: r}
		proxyCache.Get(nil, pw)

	} else {
		proxy.ServeHTTP(w, r)
	}
}

func main() {

	proxy.Director = director

	http.HandleFunc("/", serveRequest)

	http.ListenAndServe(":8080", nil)
}
