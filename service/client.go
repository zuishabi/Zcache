package service

import (
	"bytes"
	"cache"
	"cache/cachepb/cachepb"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"net/http"
)

type Client struct {
	cache.HttpGetter
}

func NewClient(URL string) *Client {
	return &Client{cache.HttpGetter{BaseURL: URL}}
}

// Get 向缓存读取数据
func (c *Client) Get(in *cachepb.GetRequest, out *cachepb.Response) error {
	return c.HttpGetter.Get(in, out)
}

// Set 向缓存设置数据
func (c *Client) Set(in *cachepb.SetRequest, out *cachepb.Response) error {
	u := fmt.Sprintf(
		"%v/%v",
		c.BaseURL,
		"GetData",
	)
	data, _ := proto.Marshal(in)
	res, err := http.Post(u, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		s, _ := io.ReadAll(res.Body)
		return fmt.Errorf("server returned: %v:%v", res.Status, string(s))
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}
	if err = proto.Unmarshal(b, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

// CreateGroup 向缓存中创建一个组
func (c *Client) CreateGroup(groupName string) {
	u := fmt.Sprintf("%v/%v?group_name=%v", c.BaseURL, "CreateGroup", groupName)
	http.Get(u)
}

// GetGroupList 获得所有组的列表
func (c *Client) GetGroupList(out *cachepb.GroupList) error {
	u := fmt.Sprintf("%v/%v", c.BaseURL, "GetGroups")
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		s, _ := io.ReadAll(res.Body)
		return fmt.Errorf("server returned: %v:%v", res.Status, string(s))
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(data, out); err != nil {
		return err
	}
	return nil
}

// GetGroupKeyList 获取一个组中的所有键
func (c *Client) GetGroupKeyList(groupName string, out *cachepb.GroupKeyList) error {
	u := fmt.Sprintf("%v/%v?group_name=%v", c.BaseURL, "GetGroupKeyList", groupName)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		s, _ := io.ReadAll(res.Body)
		return fmt.Errorf("server returned: %v:%v", res.Status, string(s))
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(data, out); err != nil {
		return err
	}
	return nil
}
