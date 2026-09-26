package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func (s *Server) Start() {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("监听失败", err)
		return
	}
	defer listen.Close()
	//启动广播goroutine
	go s.broadCast()

	for {
		accept, err := listen.Accept()
		if err != nil {
			fmt.Println("接听失败", err)
			continue
		}
		go s.handler(accept)
	}
}

// 向所有用户广播
func (s *Server) broadCast() {
	for {
		msg := <-s.Message
		s.mapLock.RLock()
		for _, v := range s.OnlineMap {
			v.C <- msg
		}
		s.mapLock.RUnlock()
	}
}

func (s *Server) handler(conn net.Conn) {
	//用户上线处理
	user := newUser(conn, s)
	user.online()

	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		//用户下线处理
		if n == 0 {
			user.offline()
			return
		}
		if err != nil && err != io.EOF {
			fmt.Println("读取出错", err)
			user.offline()
			return
		}
		msg := string(buf[:n-1])
		user.processMessage(msg)
	}
}

func newServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}
