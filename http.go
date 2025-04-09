package cache

import (
	"cache/cachepb/cachepb"
	"cache/consistenthash"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

//提供被其他节点访问的能力(基于http)

const (
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string //用来记录自己的地址，包括主机名/IP 和端口
	mu          sync.Mutex
	peers       *consistenthash.Map    // 保存其他的peer节点，根据具体的key选择peer
	httpGetters map[string]*HttpGetter // keyed by e.g. "http://10.0.0.2:8008"
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self: self,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	p.Log("%s %s", r.Method, r.URL.Path)
	parts := strings.SplitN(r.URL.Path, "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	q := r.URL.Query()
	data, _ := io.ReadAll(r.Body)
	method := parts[1]
	// 创建一个新的组
	switch method {
	case "CreateGroup":
		groupName := q.Get("group_name")
		fmt.Println("create group ", groupName)
		NewGroup(groupName, 2048, nil)
		return
	case "GetGroups":
		list := GetGroupList()
		d, err := proto.Marshal(&cachepb.GroupList{GroupName: list})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(d)
		return
	case "GetGroupKeyList":
		groupName := q.Get("group_name")
		group := GetGroup(groupName)
		if group == nil {
			http.Error(w, "no such group: "+groupName, http.StatusNotFound)
			return
		}
		d, err := proto.Marshal(&cachepb.GroupKeyList{Key: group.GetGroupKeyList()})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(d)
		return
	}
	switch r.Method {
	case "GET":
		groupName := q.Get("group")
		key := q.Get("key")
		group := GetGroup(groupName)
		if group == nil {
			http.Error(w, "no such group: "+groupName, http.StatusNotFound)
			return
		}
		view, err := group.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		body, err := proto.Marshal(&cachepb.Response{Value: view.ByteSlice()})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(body)
	case "POST":
		req := cachepb.SetRequest{}
		_ = proto.Unmarshal(data, &req)
		group := GetGroup(req.Group)
		if group == nil {
			http.Error(w, "no such group: "+req.Group, http.StatusNotFound)
			return
		}
		group.Set(req.Key, req.Value)
		body, err := proto.Marshal(&cachepb.Response{Value: []byte("create success")})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(body)
	}
}

// Set 实例化了一致性哈希算法，并且添加了传入的节点
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*HttpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &HttpGetter{BaseURL: peer}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

// HttpGetter http客户端，实现了PeerGetter接口
type HttpGetter struct {
	BaseURL string
}

func (h *HttpGetter) Get(in *cachepb.GetRequest, out *cachepb.Response) error {
	u := fmt.Sprintf(
		"%v/%v/?key=%v&group=%v",
		h.BaseURL,
		"GetData",
		url.QueryEscape(in.GetKey()),
		url.QueryEscape(in.GetGroup()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		s, _ := io.ReadAll(res.Body)
		return fmt.Errorf("server returned: %v:%v", res.Status, string(s))
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

var _ PeerGetter = (*HttpGetter)(nil)
