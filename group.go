package cache

import (
	"cache/cachepb/cachepb"
	"cache/singleflight"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"sync"
)

//负责与外部交互，控制缓存存储和获取的主流程

type groupInfo struct {
	Name       []string `yaml:"name"`
	CacheBytes []int64  `yaml:"cache-bytes"`
}

// LoadGroups 加载组文件，将组信息导入
func LoadGroups() {
	fmt.Println("loading groups info...")
	f, err := os.OpenFile("groups.yml", os.O_RDWR, 0644)
	if err != nil {
		//未找到对应文件，创建新文件
		f, err = os.Create("groups.yml")
		if err != nil {
			panic(err)
		}
		data, err := yaml.Marshal(groupInfo{
			Name:       []string{"default"},
			CacheBytes: []int64{2048},
		})
		if err != nil {
			panic(err)
		}
		if _, err := f.Write(data); err != nil {
			panic(err)
		}
	}
	data, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	g := groupInfo{}
	if err := yaml.Unmarshal(data, &g); err != nil {
		panic(err)
	}
	if len(g.Name) != len(g.CacheBytes) {
		panic(errors.New("wrong groups file"))
	}
	for i := range len(g.Name) {
		NewGroup(g.Name[i], g.CacheBytes[i], nil)
	}
	defer f.Close()
}

func UpdateGroupInfo() {
	mu.Lock()
	defer mu.Unlock()
	g := groupInfo{
		Name:       make([]string, len(groups)),
		CacheBytes: make([]int64, len(groups)),
	}
	i := 0
	for _, v := range groups {
		g.Name[i] = v.name
		g.CacheBytes[i] = v.mainCache.cacheBytes
		i += 1
	}
	f, err := os.OpenFile("groups.yml", os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	data, err := yaml.Marshal(g)
	if err != nil {
		fmt.Println(err)
		return
	}
	if _, err := f.Write(data); err != nil {
		fmt.Println(err)
	}
}

// Getter 获取对应key的数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 接口型函数
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter //用户设定的getter，在找不到对应数据时调用此回调函数在本地数据库中进行查找
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group //用来防止缓存穿透
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// RegisterPeers 注册一个用来选择远程节点的方法,将 实现了 PeerPicker 接口的 HTTPPool 注入到 Group 中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// Get 获取数据
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}
	// 如果在缓存中没有找到对应的数据，则从本地获取，通过用户设置的回调函数
	return g.load(key)
}

// 当本地缓存中未命中时，会先尝试从远程节点中获取，如果也没有，则通过本地用户设置的回调函数中获取
func (g *Group) load(key string) (value ByteView, err error) {
	//确保只会调用一次
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// 从远程节点获得对应组的数据
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &cachepb.GetRequest{
		Group: g.name,
		Key:   key,
	}
	res := &cachepb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

// 通过用户设定的getter函数从源数据中获得数据，并放入缓存中
func (g *Group) getLocally(key string) (ByteView, error) {
	if g.getter == nil {
		return ByteView{}, errors.New("FoundNoData")
	}
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// Set 设置数据
func (g *Group) Set(key string, value ByteView) {
	//TODO 设置分布式节点的设置数据
	g.mainCache.add(key, value)
}

// GetGroupKeyList 获得一个组中所有的键
func (g *Group) GetGroupKeyList() []string {
	return g.mainCache.getKeyList()
}

// SaveGroup 将组中的数据进行数据持久化
func (g *Group) SaveGroup(f *os.File) error {
	info := GroupInfo{}
	info.Name = g.name
	return g.mainCache.saveCache(f, &info)
}

// Delete 删除组中所对应的键值，通过返回一个布尔值获取是否成功删除
func (g *Group) Delete(key string) bool {
	return g.mainCache.delete(key)
}

func (g *Group) SetList(keys []string, values []ByteView) {
	g.mainCache.addList(keys, values)
}

//---------------------------------------------------------------------------------------------------------------------

// GetGroupList 获得group列表
func GetGroupList() []string {
	mu.Lock()
	defer mu.Unlock()
	res := make([]string, 0)
	for i, _ := range groups {
		res = append(res, i)
	}
	return res
}

// DeleteGroup 删除组中的所有内容
func DeleteGroup(groupName string) {
	mu.Lock()
	defer mu.Unlock()
	delete(groups, groupName)
}
