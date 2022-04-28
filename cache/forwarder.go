package groupcache

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/golang/groupcache/cache/consistenthash"
)

const DEFAULT_REPLICAS = 50

type forwarder struct {
	self string

	mu    sync.Mutex
	peers *consistenthash.Map
}

func newForwarder() *forwarder {
	return &forwarder{}
}

func (f *forwarder) Forward(key string, r *http.Request) (res *http.Response, err error, ok bool) {

	peer, ok := f.pickPeer(key)
	if !ok {
		return
	}

	bodyAsBytes, err := getRequestBodyAsBytes(r)
	if err != nil {
		return
	}

	// TODO do more?
	freq, err := http.NewRequest(r.Method, peer, bytes.NewReader(*bodyAsBytes))

	freq.Header = make(http.Header)
	for h, val := range r.Header {
		freq.Header[h] = val
	}

	tr := http.DefaultTransport
	res, err = tr.RoundTrip(freq)
	return
}

func (f *forwarder) set(self string, peers ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.self = self
	f.peers = consistenthash.New(DEFAULT_REPLICAS, nil)
	f.peers.Add(peers...)
}

func (f *forwarder) pickPeer(key string) (peer string, ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.peers.IsEmpty() {
		return
	}
	if peer = f.peers.Get(key); peer != f.self {
		return peer, true
	}
	return
}

/*

func forward(w http.ResponseWriter, req *http.Request, peer string) {
	// we need to buffer the body if we want to read it here and send it
	// in the request.
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// you can reassign the body if you need to parse it as multipart
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	// create a new url from the raw RequestURI sent by the client
	url := fmt.Sprintf("%s://%s%s", "http", peer, req.RequestURI)
	fmt.Println(url)

	proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))

	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = req.Header
	proxyReq.Header = make(http.Header)
	for h, val := range req.Header {
		proxyReq.Header[h] = val
	}
	proxyReq.Header.Set("Picked-By-Peer", "True")
	//	proxyReq.Header.Set("Target-Port", ":8081")

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))

	// legacy code
}


*/
