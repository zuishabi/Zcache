package cache

//缓存值的抽象与封装

// ByteView 表示缓存值，是一个只读的数据结构
type ByteView struct {
	b []byte
}

func (b ByteView) Len() int {
	return len(b.b)
}

// ByteSlice 返回一个拷贝，放置缓存值被外部程序修改
func (b ByteView) ByteSlice() []byte {
	return cloneBytes(b.b)
}

// String 返回数据的string值
func (b ByteView) String() string {
	return string(b.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
