package operator

import "container/heap"

type GDSF struct {
	MaxEntries int
	OnEvicted  func(key Key, value interface{})

	L float64

	pq    PriorityQueue
	cache map[interface{}]*Item
}

func NewGDSF(maxEntries int) *GDSF {
	return &GDSF{
		MaxEntries: maxEntries,
		L:          0,
		pq:         make(PriorityQueue, 0),
		cache:      make(map[interface{}]*Item),
	}
}

func (g *GDSF) Add(key Key, value interface{}) {

	if it, ok := g.cache[key]; ok {
		// TODO update priority
		// calculate priority
		g.pq.update(it, value, it.priority)
	} else {
		priority := 1
		it := Item{
			key:   key,
			value: value,
			// TODO
			priority: float64(priority),
		}
		heap.Push(&g.pq, it)
		g.pq.update(&it, value, float64(priority))
		g.cache[key] = &it
	}
}
