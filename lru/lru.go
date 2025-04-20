// Package lru 缓存淘汰策略
package lru

import (
	"container/list"
	"fmt"
)

type Cache struct {
	maxBytes int64 //允许使用的最大内存
	nbytes   int64 //当前使用的内存
	ll       *list.List
	cache    map[string]*list.Element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

type Entry struct {
	Key   string
	Value Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

// New 创建缓存
func New(maxByte int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxByte,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get 查找缓存中的元素，首先从字典中找到双向链表中对应的节点，然后将节点移动到队首
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*Entry)
		return kv.Value, true
	}
	fmt.Println(c.cache)
	return
}

// RemoveOldest 缓存淘汰，淘汰最近最少访问的节点
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*Entry)
		delete(c.cache, kv.Key)
		c.nbytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
		//如果回调函数不为nil，则调用回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.Key, kv.Value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*Entry)
		c.nbytes += int64(value.Len()) - int64(kv.Value.Len())
		kv.Value = value
	} else {
		ele := c.ll.PushFront(&Entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Delete 删除对应键的值
func (c *Cache) Delete(key string) bool {
	ele := c.cache[key]
	if ele == nil {
		return false
	}
	delete(c.cache, key)
	c.ll.Remove(ele)
	c.nbytes += int64(ele.Value.(*Entry).Value.Len()) + int64(len(key))
	return true
}

// Len the number of cache entries
func (c *Cache) Len() int {
	if c == nil {
		return 0
	}
	return c.ll.Len()
}

// GetKeyList 获得所有键的列表
func (c *Cache) GetKeyList() []string {
	if c == nil {
		return nil
	}
	res := make([]string, 0)
	for _, v := range c.cache {
		res = append(res, v.Value.(*Entry).Key)
	}
	return res
}

// GetKVList 用于返回当前内存中所有的键值对，用于快速批量获取缓存内容
func (c *Cache) GetKVList() []Entry {
	res := make([]Entry, c.Len())
	i := 0
	for e := c.ll.Front(); e != nil; e = e.Next() {
		res[i] = *e.Value.(*Entry)
		i++
	}
	return res
}
