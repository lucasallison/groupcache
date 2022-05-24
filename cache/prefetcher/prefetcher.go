package prefetcher

import (
	"fmt"

	"github.com/golang/groupcache/cache/operator"
)

/* prefetcher configurations */
const MAX_TRIES int = 1
const MAX_TRACE_SIZE int = 5
const MIN_URI_MATCHES int = 2

type Prefetcher struct {
	tries *operator.LRU
}

func NewPrefetcher() *Prefetcher {
	pf := &Prefetcher{
		tries: operator.NewLRU(MAX_TRIES, nil),
	}

	pf.tries.OnEvicted = func(key operator.Key, value interface{}) {

		fmt.Println("Evicting user ", key, " trie")
		trie, ok := value.(*Trie)
		if !ok {
			return
		}
		// TODO delete file?
		trie.SaveTrie()
	}

	createBaseDir()

	return pf
}

/* Adds the request to the user trace in the trie. Returns a list of predictions */
func (tm *Prefetcher) ProcessRequest(uid string, URI string) []string {

	var trie *Trie
	/* check if the users trie is in memory */
	value, ok := tm.tries.Get(uid)

	/* load users trie into memory */
	if !ok {
		fmt.Println("TRIE made")
		trie = NewTrie(uid)
		tm.tries.Add(uid, trie, 0)
	} else {
		fmt.Println("TRIE found")
		trie = value.(*Trie)
		fmt.Println(trie.uid)
	}

	trie.ProcessRequest(URI)

	return nil
}
