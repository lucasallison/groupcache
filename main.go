package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	groupcache "github.com/golang/groupcache/cache"
	"github.com/golang/groupcache/cache/prefetcher"
	"github.com/golang/groupcache/utils"
)

// var cacheBytes int64 = 20000

var cacheBytes int64 = 64 << 20
var cacheOperator string = "GDSF"
var validation bool = false
var admission bool = false
var logsEnabled bool = true

// TODO fix this
var groupName string = "a"

var proxyCache *groupcache.ProxyCache
var pf = prefetcher.NewPrefetcher()
var prefetchingEnabled bool = utils.PrefetchingEnabled()
var host string = utils.GetHostFromEnv()

// just for logging
var addr string
var peers string

func setProxyCache() {
	proxyCache = groupcache.NewProxyCache(cacheBytes, validation, cacheOperator, admission, logsEnabled, groupName)
	p := strings.Split(peers, ",")
	proxyCache.RegisterPeerGroup(p[0], p...)
}

func parsePatchRequest(w http.ResponseWriter, path string) {

	parts := strings.Split(path, "/")
	parts = parts[1:]
	cb, err := strconv.Atoi(parts[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cacheBytes = int64(cb)

	cacheOperator = parts[1]
	validation = stringToBool(parts[2])
	admission = stringToBool(parts[3])
	logsEnabled = stringToBool(parts[4])

	groupName += "a"

	setProxyCache()
	w.WriteHeader(http.StatusOK)
}

func stringToBool(s string) bool {
	if s == "true" {
		return true
	}
	return false
}

func serveRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPatch {
		parsePatchRequest(w, r.URL.Path)
		return
	}

	var proxy = httputil.NewSingleHostReverseProxy(&url.URL{})
	proxy.Director = func(r *http.Request) {
		r.Host = host
		r.URL.Host = r.Host
		r.URL.Scheme = "http"
	}

	if r.URL.Path == "/control/cache/stats" {
		stats := proxyCache.LogStats()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(stats))
		return
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

	addr, ok := utils.GetEnvVariable("addr")
	if !ok {
		log.Fatal("addr not in env file")
	}
	peers, ok = utils.GetEnvVariable("peers")
	if !ok {
		log.Fatal("addr not in env file")
	}

	setProxyCache()
	http.HandleFunc("/", serveRequest)

	log.Println("Servering at: http://localhost" + addr)
	err := http.ListenAndServe(addr, nil)
	log.Println("ERROR for cache on port ", addr, ": ", err.Error())
}
