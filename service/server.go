package service

import (
	"cache"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

/*
提供默认的server，在指定的端口进行服务
提供默认的api服务
*/

type Server struct {
	ip              string //服务器的ip地址
	port            int    //服务器的端口
	persistence     bool   //是否开启持久化
	persistenceTime int    // 数据持久化的时间
}

func NewServer(ip string, port int, persistence bool, persistenceTime int) *Server {
	return &Server{
		ip:              ip,
		port:            port,
		persistence:     persistence,
		persistenceTime: persistenceTime,
	}
}

func (s *Server) Run() {
	cache.NewGroup("default", 2048, nil)
	addr := s.ip + ":" + strconv.Itoa(s.port)
	pool := cache.NewHTTPPool(addr)
	wg := sync.WaitGroup{}
	//加载组文件
	cache.LoadGroups()
	if s.persistence {
		cache.LoadPersistence()
		go s.savePersistence(&wg)
	}
	go ListenSignal(&wg)
	log.Fatal(http.ListenAndServe(addr, pool))
}

// 进行持久化工作
func (s *Server) savePersistence(wg *sync.WaitGroup) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case <-c:
			//当程序退出时再进行一次保存
			cache.SavePersistence()
			return
		case <-time.After(time.Second * time.Duration(s.persistenceTime)):
			//当倒计时结束时进行一次保存
			cache.SavePersistence()
		}
	}
}

// ListenSignal 监听信号2，15，当收到信号时关闭reader
func ListenSignal(wg *sync.WaitGroup) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	cache.UpdateGroupInfo()
	wg.Wait()
	fmt.Println("bye bye~")
	os.Exit(0)
}
