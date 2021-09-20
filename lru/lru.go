package lru

import "container/list"

type Cache struct {
	maxBytes   int64
	usedBytes  int64
	linkedList *list.List
	cacheMap   map[string]*list.Element
	OnEvicted  func(key string, val Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func NewCache(maxBytes int64, onEvicted func(key string, val Value)) *Cache {
	return &Cache{
		maxBytes:   maxBytes,
		linkedList: list.New(),
		cacheMap:   make(map[string]*list.Element),
		OnEvicted:  onEvicted,
	}
}

func (c *Cache) Get(key string) (Value, bool) {
	// 1. get the element (node in doubly linked list from the map)
	// 2. move the node to the front
	if ele, ok := c.cacheMap[key]; ok {
		c.linkedList.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, ok
	}

	return nil, false
}

func (c *Cache) RemoveOldest() {
	ele := c.linkedList.Back()

	if ele != nil {
		c.linkedList.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cacheMap, kv.key)

		c.usedBytes -= int64(len(kv.key)) + int64(kv.value.Len())

		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	// if the key exists, update it and move to the front
	// if not, add it
	if ele, ok := c.cacheMap[key]; ok {
		c.linkedList.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.usedBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.linkedList.PushFront(&entry{
			key:   key,
			value: value,
		})
		c.cacheMap[key] = ele
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.usedBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.linkedList.Len()
}
