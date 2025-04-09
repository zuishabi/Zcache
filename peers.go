package cache

import "cache/cachepb/cachepb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) //用于根据传入的key选择相应的peergetter节点
}

type PeerGetter interface {
	Get(in *cachepb.GetRequest, out *cachepb.Response) error //从对应的perrgetter中获取对应group中的对应key的值
}
