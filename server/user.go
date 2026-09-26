package main

import (
	"fmt"
	"net"
)

type User struct {
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func newUser(conn net.Conn, server *Server) *User {
	u := &User{
		Addr:   conn.RemoteAddr().String(),
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go u.sendMessage()
	return u
}

func (u *User) online() {
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Addr] = u
	u.server.mapLock.Unlock()
	u.server.Message <- fmt.Sprintf("[%s]上线了", u.Addr)
}
func (u *User) offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Addr)
	u.server.mapLock.Unlock()
	u.server.Message <- fmt.Sprintf("[%s]下线了", u.Addr)
}

// 处理用户发来的消息
func (u *User) doMessage(msg string) {
	//向所有用户广播
	u.server.Message <- msg
}

// 向用户发送消息
func (u *User) sendMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
