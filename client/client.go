package main

import (
	"fmt"
	"net"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	Conn       net.Conn
}

func newClient(serverIP string, serverPort int) *Client {
	dial, err := net.Dial("tcp", net.JoinHostPort(serverIP, fmt.Sprintf("%d", serverPort)))
	if err != nil {
		fmt.Println("「客户端」: 连接失败", err)
	}
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
		Name:       "",
		Conn:       dial,
	}
	return client
}
