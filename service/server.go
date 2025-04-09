package service

import (
	"cache"
	"log"
	"net/http"
	"strconv"
)

/*
提供默认的server，在指定的端口进行服务
提供默认的api服务
*/

type Server struct {
	ip   string //服务器的ip地址
	port int    //服务器的端口
}

func NewServer(ip string, port int) *Server {
	return &Server{ip: ip, port: port}
}

func (s *Server) Run() {
	cache.NewGroup("default", 2048, nil)
	addr := s.ip + ":" + strconv.Itoa(s.port)
	pool := cache.NewHTTPPool(addr)
	log.Fatal(http.ListenAndServe(addr, pool))
}
