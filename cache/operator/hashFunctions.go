package operator

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/binary"
	"hash/fnv"
)

func hashFNV(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

func hashSHA1(s string) int {
	h := sha1.New()
	h.Write([]byte(s))
	hash := h.Sum(nil)
	return int(binary.BigEndian.Uint64(hash))
}

func hashMD5(s string) int {
	h := md5.New()
	h.Write([]byte(s))
	hash := h.Sum(nil)
	return abs(int(binary.BigEndian.Uint64(hash)))
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
