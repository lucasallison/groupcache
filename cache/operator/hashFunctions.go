package operator

import (
	"bytes"
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
	var hash int
	hashBytes := sha1.Sum([]byte(s))
	buf := bytes.NewBuffer(hashBytes[:])
	binary.Read(buf, binary.LittleEndian, &hash)
	return hash
}

func hashMD5(s string) int {
	var hash int
	hashBytes := md5.Sum([]byte(s))
	buf := bytes.NewBuffer(hashBytes[:])
	binary.Read(buf, binary.LittleEndian, &hash)
	return hash
}
