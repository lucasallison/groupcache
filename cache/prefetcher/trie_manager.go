package prefetcher

import "container/list"

// TODO LRU; restrict memory
type TrieManger struct {
	tries *list.List
	users map[string]*list.Element
}

func NewTrieManger() *TrieManger {
	return nil
}

func (tm *TrieManger) getUserTrie(uid string) *list.Element {
	if ee, ok := tm.users[uid]; ok {
		return ee
	}
	return nil

}
