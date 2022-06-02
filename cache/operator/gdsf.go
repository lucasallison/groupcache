package operator

import (
	"container/heap"
)

type GDSF struct {
	MaxEntries int
	OnEvicted  func(key Key, value interface{})

	L float64

	pq    PriorityQueue
	cache map[interface{}]*Item
}

func NewGDSF(maxEntries int, onEvicted func(key Key, value interface{})) *GDSF {
	return &GDSF{
		MaxEntries: maxEntries,
		L:          0,
		pq:         make(PriorityQueue, 0),
		cache:      make(map[interface{}]*Item),
		OnEvicted:  onEvicted,
	}
}

func (g *GDSF) Add(key Key, value interface{}, len int) {

	priority := len
	if it, ok := g.cache[key]; ok {
		g.pq.update(it, value, float64(priority))
	} else {
		it := Item{
			key:      key,
			value:    value,
			len:      len,
			priority: float64(priority),
		}
		heap.Push(&g.pq, &it)
		g.pq.update(&it, value, float64(priority))
		g.cache[key] = &it
	}
}

func (g *GDSF) Get(key Key) (value interface{}, ok bool) {
	if ele, hit := g.cache[key]; hit {
		// TODO update PRIORITY!
		return ele.value, true
	}
	return
}

func (g *GDSF) Remove(key Key) {
	if _, hit := g.cache[key]; hit {
		g.RemoveBasedOnPolicy()
	}
}

func (g *GDSF) RemoveBasedOnPolicy() {

	item := heap.Pop(&g.pq).(*Item)
	delete(g.cache, item.key)

	if g.OnEvicted != nil {
		g.OnEvicted(item.key, item.value)
	}

}

func (g *GDSF) Len() int {
	return g.pq.Len()
}

func (g *GDSF) Clear() {

	for g.pq.Len() > 0 {
		item := heap.Pop(&g.pq).(*Item)
		delete(g.cache, item.key)
	}
}

func (g *GDSF) ContainsKey(key Key) bool {
	_, ok := g.cache[key]
	return ok
}

func (g *GDSF) NextVictim() (key string) {
	item := heap.Pop(&g.pq).(*Item)
	if item == nil {
		return
	}

	g.Add(item.key, item.value, item.len)
	return item.key.(string)
}
