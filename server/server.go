package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type Server struct {
	IP        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func (s *Server) Start() {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
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
	for msg := range s.Message {
		s.mapLock.RLock()
		users := make([]*User, 0, len(s.OnlineMap))
		for _, user := range s.OnlineMap {
			users = append(users, user)
		}
		s.mapLock.RUnlock()

		for _, user := range users {
			user.deliver(msg)
		}
	}
}

func (s *Server) handler(conn net.Conn) {
	//用户上线处理
	user := newUser(conn, s)
	user.online()

	defer func() {
		conn.Close()
		close(user.done)
		user.offline()

	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		user.processMessage(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("读取出错：", err)
	}
}

func newServer(ip string, port int) *Server {
	return &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}
