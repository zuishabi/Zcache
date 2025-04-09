package service

import (
	"cache/cachepb/cachepb"
	"fmt"
	"testing"
)

func TestClient(t *testing.T) {
	s := NewServer("127.0.0.1", 8999)
	go s.Run()
	c := NewClient("http://127.0.0.1:8999")
	in := cachepb.GetRequest{Group: "a", Key: "a"}
	out := &cachepb.Response{}
	if err := c.Get(&in, out); err != nil {
		fmt.Println("get error", err)
	}
	set := cachepb.SetRequest{
		Group: "a",
		Key:   "a",
		Value: []byte("111"),
	}
	c.CreateGroup("a")
	if err := c.Set(&set, out); err != nil {
		fmt.Println("set error", err)
	}
	fmt.Println(string(out.Value))
	if err := c.Get(&in, out); err != nil {
		fmt.Println("get error", err)
	}
	fmt.Println(string(out.Value))
	select {}
}
