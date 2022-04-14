package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	groupcache "github.com/golang/groupcache/cache"
	"github.com/golang/groupcache/cache/prefetcher"
)

//var cacheBytes int64 = 64 << 20
var cacheBytes int64 = 1500
var proxyCache = groupcache.NewProxyCache(cacheBytes)
var pf = prefetcher.NewPrefetcher()
var prefetchingEnabled bool = PrefetchingEnabled()
var proxy = httputil.NewSingleHostReverseProxy(&url.URL{})

func director(r *http.Request) {
	r.Host = GetHostFromEnv()
	r.URL.Host = r.Host
	r.URL.Scheme = "http"
}

var uid string = "1"

func serveRequest(w http.ResponseWriter, r *http.Request) {

	// TODO remove
	if r.URL.Path == "/switch" {
		if uid == "1" {
			uid = "2"
		} else {
			uid = "1"
		}
		fmt.Println("User is now: ", uid)

		// will fail
		proxy.ServeHTTP(w, r)
		return
	}

	if r.Method == http.MethodGet {

		if prefetchingEnabled {
			pf.ProcessRequest(uid, r.URL.Path)
		}

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
