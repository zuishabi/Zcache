package cache

import (
	"cache/lru"
	"os"
	"strconv"
	"sync"
)

// 并发控制
type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}

func (c *cache) getKeyList() []string {
	mu.Lock()
	defer mu.Unlock()
	return c.lru.GetKeyList()
}

func (c *cache) saveCache(f *os.File) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := f.WriteString(strconv.Itoa(c.lru.Len()) + "\n"); err != nil {
		return err
	}
	for _, v := range c.lru.GetKeyList() {
		t, _ := c.lru.Get(v)
		if _, err := f.WriteString(v + " " + t.(ByteView).String() + "\n"); err != nil {
			return err
		}
	}
	return nil
}
