package groupcache

import (
	"bytes"
	"fmt"
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
	fmt.Println(peer)

	bodyAsBytes, err := getRequestBodyAsBytes(r)
	if err != nil {
		return
	}

	url := peer + r.URL.Path
	freq, err := http.NewRequest(r.Method, url, bytes.NewReader(*bodyAsBytes))

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
