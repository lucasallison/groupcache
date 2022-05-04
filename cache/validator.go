package groupcache

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/golang/groupcache/utils"
)

type validator struct {
	res *http.Response
	url string
}

func (v *validator) validate(r *http.Response) {
	servedResponseBody, _ := getBodyAsBytes(&r.Body)
	correctResponseBody, _ := getBodyAsBytes(&v.res.Body)

	result := bytes.Compare(*servedResponseBody, *correctResponseBody)
	if result != 0 || v.res.StatusCode != r.StatusCode {
		log.Println("INVALID response has been served for ", v.url)
	}
}

func (v *validator) makeRequest(r *http.Request) {

	bodyAsBytes, err := getBodyAsBytes(&r.Body)
	if err != nil {
		return
	}

	host := utils.GetHostFromEnv()
	v.url = fmt.Sprintf("http://%s%s", host, r.URL.Path)
	req, err := http.NewRequest(r.Method, v.url, bytes.NewReader(*bodyAsBytes))

	req.Header = make(http.Header)
	for h, val := range r.Header {
		req.Header[h] = val
	}

	tr := http.DefaultTransport
	v.res, _ = tr.RoundTrip(req)
	return
}
