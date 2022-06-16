package operator

import (
	"container/heap"
	"fmt"
)

type GDSF struct {
	MaxEntries int
	OnEvicted  func(key Key, value interface{})

	L float64

	// GDS or GDSF
	frequency bool
	pq        PriorityQueue
	cache     map[interface{}]*Item
}

func NewGDSF(maxEntries int, frequency bool, onEvicted func(key Key, value interface{})) *GDSF {
	return &GDSF{
		MaxEntries: maxEntries,
		L:          0,
		frequency:  frequency,
		pq:         make(PriorityQueue, 0),
		cache:      make(map[interface{}]*Item),
		OnEvicted:  onEvicted,
	}
}

func (g *GDSF) Add(key Key, value interface{}, len int) {

	var priority float64

	if it, ok := g.cache[key]; ok {
		priority = g.calcPriority(it.frequency, len)
		g.L = priority
		it.frequency++
		g.pq.update(it, value, float64(priority))
	} else {
		fmt.Println("Add")
		priority = g.calcPriority(1, len)
		g.L = priority
		it := Item{
			key:       key,
			value:     value,
			len:       len,
			frequency: 1,
			priority:  float64(priority),
		}
		heap.Push(&g.pq, &it)
		g.pq.update(&it, value, float64(priority))
		g.cache[key] = &it
	}
}

func (g *GDSF) calcPriority(frequency int, len int) float64 {
	if len == 0 {
		len++
	}

	fmt.Println("L: ", g.L, " freq: ", frequency, " len: ", len)

	fmt.Println(float64(frequency * (1 / len)))
	if g.frequency {
		return g.L + float64(frequency*(1/len))
	}

	return g.L + float64(frequency*(1/len))
}

func (g *GDSF) Get(key Key) (value interface{}, ok bool) {
	if ele, hit := g.cache[key]; hit {
		priority := g.calcPriority(ele.frequency, ele.len)
		g.L = priority
		ele.frequency++
		g.pq.update(ele, ele.value, float64(priority))
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
