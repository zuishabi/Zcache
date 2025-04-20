package cache

import (
	"cache/lru"
	"encoding/json"
	"os"
	"sync"
)

// 用于并发控制
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

// 用于快速批量添加数据，减少了锁的获取与释放
func (c *cache) addList(keys []string, values []ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	for i, _ := range keys {
		c.lru.Add(keys[i], values[i])
	}
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

// 获取所有的键列表
func (c *cache) getKeyList() []string {
	mu.Lock()
	defer mu.Unlock()
	if c.lru == nil {
		return nil
	}
	return c.lru.GetKeyList()
}

// 通过json序列化来将缓存中的数据进行持久化保存
func (c *cache) saveCache(f *os.File, info *GroupInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := json.NewEncoder(f)
	if c.lru == nil {
		return nil
	}
	list := c.lru.GetKVList()
	info.Num = len(list)
	if err := e.Encode(info); err != nil {
		return err
	}
	for _, v := range list {
		b := v.Value.(ByteView)
		p := ByteToPersistence(v.Key, &b)
		if err := e.Encode(&p); err != nil {
			return err
		}
	}
	return nil
}

func (c *cache) delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return false
	}
	return c.lru.Delete(key)
}
