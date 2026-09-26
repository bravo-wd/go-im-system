package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

func (s *Server) Start() {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("监听失败", err)
		return
	}
	defer listen.Close()
	for {
		accept, err := listen.Accept()
		if err != nil {
			fmt.Println("接听失败", err)
			continue
		}
		go s.handler(accept)
	}
}

func (s *Server) handler(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("读取出错", err)
			return
		}
		fmt.Println(string(buf[:n]))
	}
}

func newServer(ip string, port int) *Server {
	return &Server{
		Ip:   ip,
		Port: port,
	}
}
