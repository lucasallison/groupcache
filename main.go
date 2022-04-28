package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	groupcache "github.com/golang/groupcache/cache"
	"github.com/golang/groupcache/cache/prefetcher"
)

var cacheBytes int64 = 64 << 20

// var cacheBytes int64 = 1500
var proxyCache = groupcache.NewProxyCache(cacheBytes)
var pf = prefetcher.NewPrefetcher()
var prefetchingEnabled bool = PrefetchingEnabled()

var proxy = httputil.NewSingleHostReverseProxy(&url.URL{})

func director(r *http.Request) {
	r.Host = GetHostFromEnv()
	r.URL.Host = r.Host
	r.URL.Scheme = "http"
}

func serveRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		/*
			if prefetchingEnabled {
				pf.ProcessRequest(uid, r.URL.Path)
			}
		*/

		/* Check the proxy cache for GET requests */
		pw := groupcache.ProxyWrapper{Proxy: proxy, Writer: w, Req: r}
		proxyCache.Get(nil, pw)

	} else {
		proxy.ServeHTTP(w, r)
	}
}

func main() {

	addr := flag.String("addr", ":8080", "server address")
	peers := flag.String("pool", "http://localhost:8080", "server pool list")
	flag.Parse()

	proxy.Director = director

	p := strings.Split(*peers, ",")
	proxyCache.RegisterPeerGroup(p[0], p...)

	http.HandleFunc("/", serveRequest)

	log.Println("Servering at: http://localhost" + *addr)
	http.ListenAndServe(*addr, nil)
}
