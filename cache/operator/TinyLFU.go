package operator

import (
	"fmt"
	"sync"
)

/* TINYLFU is not a cache operator, but always works in combintion with one */
const FILTERSIZE int = 10000
const RESETFREQUENCY int = 100

// TODO reset completely when a certain number is hit?
// TODO gatekeeping?

type TinyLFU struct {
	mu            sync.RWMutex
	bloomFilter   [FILTERSIZE]int
	hashFuncs     []func(s string) int
	accessCounter int
}

func NewTinyLFU() *TinyLFU {

	t := TinyLFU{}

	/* Add desired hash functions */
	t.hashFuncs = append(t.hashFuncs, hashFNV)
	t.hashFuncs = append(t.hashFuncs, hashSHA1)
	t.hashFuncs = append(t.hashFuncs, hashMD5)

	return &t
}

func (t *TinyLFU) RegisterEntry(key string) {

	t.accessCounter = t.accessCounter % RESETFREQUENCY
	if t.accessCounter == 0 {
		t.Reset()
	}
	t.accessCounter++

	t.mu.Lock()
	defer t.mu.Unlock()

	first := true
	var highest int
	for _, hash := range t.hashFuncs {
		i := hash(key) % FILTERSIZE
		fmt.Println("hash: ", i)
		t.bloomFilter[i]++
		if first {
			highest = i
			first = false
		} else if t.bloomFilter[highest] < t.bloomFilter[i] {
			highest = i
		}
	}

	t.bloomFilter[highest]--
}

func (t *TinyLFU) EstimateFrequency(key string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	sum := 0
	highest := 0
	for _, hash := range t.hashFuncs {
		i := hash(key) % FILTERSIZE
		sum += t.bloomFilter[i]

		if t.bloomFilter[i] > highest {
			highest = t.bloomFilter[i]
		}
	}

	// Remove outliers that might be very large
	return (sum - highest) / (len(t.hashFuncs) - 1)
}

func (t *TinyLFU) AdmitNewKey(newKey string, existingKey string) bool {

	nfreq := t.EstimateFrequency(newKey)
	efreq := t.EstimateFrequency(existingKey)

	fmt.Println("nkey freq: ", nfreq, " ekey freq: ", efreq)
	// TODO make this more sophisticated, add some margin or something ...
	return nfreq > efreq
}

func (t *TinyLFU) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, e := range t.bloomFilter {
		e = e / 2
	}
}
