package operator

import "container/list"

type LFU struct {
	MaxEntries int

	OnEvicted func(key Key, value interface{})

	ll    *list.List
	cache map[interface{}]*list.Element
}

// TODO see if we can make this a struct fuc
/* increase access frequency by one and potentially move it up the list */
func registerAccess(ll *list.List, el *list.Element) {

	el.Value.(*entry).accessFrequency++

	/* compare neighbours acces frequency with its own */
	caf := el.Value.(*entry).accessFrequency
	for {

		if el.Next() == nil {
			return
		}

		naf := el.Next().Value.(*entry).accessFrequency
		if naf > caf {
			return
		}
		ll.MoveAfter(el, el.Next())
	}
}

func (c *LFU) Add(key Key, value interface{}) {
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	if ee, ok := c.cache[key]; ok {

		registerAccess(c.ll, ee)

		ee.Value.(*entry).value = value
		return
	}

	ele := c.ll.PushBack(&entry{key, value, 0})
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveBasedOnPolicy()
	}
}

func (c *LFU) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		registerAccess(c.ll, ele)
		return ele.Value.(*entry).value, true
	}
	return
}

func (c *LFU) Remove(key Key) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

func (c *LFU) RemoveBasedOnPolicy() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *LFU) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *LFU) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// Clear purges all stored items from the cache.
func (c *LFU) Clear() {
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.cache = nil
}
