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

// var cacheBytes int64 = 20000

// TODO env variable?
var cacheBytes int64 = 64 << 20
var cacheOperator string = "LRU"
var validation bool = true
var admission bool = false
var logsEnabled bool = false

var proxyCache = groupcache.NewProxyCache(cacheBytes, validation, cacheOperator, admission, logsEnabled)
var pf = prefetcher.NewPrefetcher()
var prefetchingEnabled bool = utils.PrefetchingEnabled()
var host string = utils.GetHostFromEnv()

func serveRequest(w http.ResponseWriter, r *http.Request) {

	var proxy = httputil.NewSingleHostReverseProxy(&url.URL{})
	proxy.Director = func(r *http.Request) {
		r.Host = host
		r.URL.Host = r.Host
		r.URL.Scheme = "http"
	}

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
		if logsEnabled {
			log.Println("Proxying a ", r.Method, " request for ", r.URL.Path)
		}
		proxy.ServeHTTP(w, r)
	}
}

// example: go run . -addr=:8080 -pool=http://127.0.0.1:8080,http://127.0.0.1:8081,http://127.0.0.1:8082
func main() {

	addr := flag.String("addr", ":8080", "server address")
	peers := flag.String("pool", "http://localhost:8080", "server pool list")
	flag.Parse()

	p := strings.Split(*peers, ",")
	proxyCache.RegisterPeerGroup(p[0], p...)

	http.HandleFunc("/", serveRequest)

	log.Println("Servering at: http://localhost" + *addr)
	http.ListenAndServe(*addr, nil)
}
