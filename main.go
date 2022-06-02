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
	"github.com/golang/groupcache/utils"
)

var cacheBytes int64 = 1000

// TODO env variable?
//var cacheBytes int64 = 64 << 20
var cacheOperator string = "LRU"
var admission bool = true

var proxyCache = groupcache.NewProxyCache(cacheBytes, true, cacheOperator, admission)
var pf = prefetcher.NewPrefetcher()
var prefetchingEnabled bool = utils.PrefetchingEnabled()

var proxy = httputil.NewSingleHostReverseProxy(&url.URL{})

func director(r *http.Request) {
	r.Host = utils.GetHostFromEnv()
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
		log.Println("Proxying a ", r.Method, " request for ", r.URL.Path)
		proxy.ServeHTTP(w, r)
	}
}

// example: go run . -addr=:8080 -pool=http://127.0.0.1:8080,http://127.0.0.1:8081,http://127.0.0.1:8082
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
