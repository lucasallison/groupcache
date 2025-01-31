/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package lru implements an LRU cache.
package operator

import (
	"container/list"
)

// LRU is an LRU cache. It is not safe for concurrent access.
type LRU struct {
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	MaxEntries int

	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted func(key Key, value interface{})

	ll    *list.List
	cache map[interface{}]*list.Element
}

// New creates a new LRU operator.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func NewLRU(maxEntries int, onEvicted func(key Key, value interface{})) *LRU {
	return &LRU{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
		OnEvicted:  onEvicted,
	}
}

// Add adds a value to the cache.
func (c *LRU) Add(key Key, value interface{}, len int) {
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	ele := c.ll.PushFront(&entry{key, value, 0})
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveBasedOnPolicy()
	}
}

// Get looks up a key's value from the cache.
func (c *LRU) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *LRU) Remove(key Key) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveBasedOnPolicy removes the oldest item from the cache.
func (c *LRU) RemoveBasedOnPolicy() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *LRU) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// Clear purges all stored items from the cache.
func (c *LRU) Clear() {
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.cache = nil
}

func (c *LRU) ContainsKey(key Key) bool {
	if c.cache == nil {
		return false
	}

	_, ok := c.cache[key]
	return ok
}

func (c *LRU) NextVictim() (key string) {
	if c.cache == nil {
		return ""
	}
	ele := c.ll.Back()
	if ele != nil {
		return ele.Value.(*entry).key.(string)
	}
	return ""
}
