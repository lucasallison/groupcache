package groupcache

import (
	"crypto/sha1"
	"sync"
)

type Validator struct {
	mu sync.RWMutex
	/* map from key to its hash */
	hashes map[string][]byte
}

func NewValidator() *Validator {
	v := Validator{}
	v.hashes = make(map[string][]byte)
	return &v
}

func (v *Validator) calcHash(b []byte) []byte {
	hasher := sha1.New()
	hasher.Write(b)
	return hasher.Sum(nil)
}

func (v *Validator) SaveEtag(URI string, value string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.hashes[URI] = v.calcHash([]byte(value))
}

func (v *Validator) getEtag(URI string) string {
	return string(v.hashes[URI])
}

/* check if the supplied hash mathes the hash associated with the key */
func (v *Validator) ValidateLocal(URI string, etag string) bool {
	return v.getEtag(URI) == etag
}
