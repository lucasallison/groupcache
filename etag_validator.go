package groupcache

import (
	"crypto/sha1"
	"sync"
)

type Etagger struct {
	mu sync.RWMutex
	/* map from key to its hash */
	hashes map[string][]byte
}

func NewTagger() *Etagger {
	v := Etagger{}
	v.hashes = make(map[string][]byte)
	return &v
}

func (v *Etagger) calcHash(b *[]byte) []byte {
	hasher := sha1.New()
	hasher.Write(*b)
	return hasher.Sum(nil)
}

func (v *Etagger) SaveBodyAsTag(URI string, b *[]byte) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.hashes[URI] = v.calcHash(b)
}

func (v *Etagger) getEtag(URI string) string {
	return string(v.hashes[URI])
}

/* check if the supplied hash mathes the hash associated with the key */
func (v *Etagger) ValidateLocal(URI string, etag string) bool {
	return v.getEtag(URI) == etag
}
