package operator

/* TINYLFU is not a cache operator, but always works in combintion with one */

const FILTERSIZE int = 10000
const RESETFREQUENCY int = 100

type TinyLFU struct {
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

	first := true
	var highest int
	for _, hash := range t.hashFuncs {
		i := hash(key) % FILTERSIZE
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
	sum := 0
	for _, hash := range t.hashFuncs {
		i := hash(key) % FILTERSIZE
		sum += t.bloomFilter[i]
	}
	return sum / len(t.hashFuncs)
}

func (t *TinyLFU) ReplaceEntry(newKey string, existingKey string) bool {
	nfreq := t.EstimateFrequency(newKey)
	efreq := t.EstimateFrequency(existingKey)

	// TODO make this more sophisticated, add some margin or something ...
	return nfreq > efreq
}

func (t *TinyLFU) Reset() {
	for _, e := range t.bloomFilter {
		e = e / 2
	}
}
