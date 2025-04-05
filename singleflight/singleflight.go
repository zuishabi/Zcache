package singleflight

import "sync"

// 表示正在进行中，或者已经结束的请求，使用waitgroup锁避免重入
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group 管理不同key的call
type Group struct {
	mu sync.Mutex //保护对m的访问
	m  map[string]*call
}

// Do 针对相同的key，无论Do被调用多少次，函数fn只会被调用一次，等待fn调用完毕，返回返回值或错误
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()
	c.val, c.err = fn()
	c.wg.Done()
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	return c.val, c.err
}
