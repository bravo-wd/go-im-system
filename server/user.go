package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func newUser(conn net.Conn, server *Server) *User {
	u := &User{
		Name:   conn.RemoteAddr().String(),
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
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()
	u.server.Message <- fmt.Sprintf("「服务器」: \"%s\"上线了", u.Name)
}
func (u *User) offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()
	u.server.Message <- fmt.Sprintf("「服务器」: \"%s\"下线了", u.Name)
}

// 处理用户发来的消息
func (u *User) processMessage(msg string) {
	switch {
	case msg == "all":
		var allUserName []string
		u.server.mapLock.RLock()
		for k := range u.server.OnlineMap {
			allUserName = append(allUserName, k)
		}
		u.server.mapLock.RUnlock()
		u.C <- fmt.Sprintf("「服务器」: ----在线用户列表----")
		for _, name := range allUserName {
			u.C <- name
		}
	case msg == "rename|":
		u.C <- fmt.Sprintf("「服务器」: 用户名不能为空")

	case len(msg) > 7 && msg[:7] == "rename|":
		_, name, _ := strings.Cut(msg, "|")
		var serverMsg string
		u.server.mapLock.Lock()
		onlineUser, exist := u.server.OnlineMap[name]
		if exist {
			if onlineUser == u {
				serverMsg = fmt.Sprintf("「服务器」: 用户名已是\"%s\"", name)
			} else {
				serverMsg = fmt.Sprintf("「服务器」: 该用户名已被使用")
			}
		} else {
			delete(u.server.OnlineMap, u.Name)
			u.Name = name
			u.server.OnlineMap[u.Name] = u
			serverMsg = fmt.Sprintf("「服务器」: 用户名更改成功")
		}
		u.server.mapLock.Unlock()
		u.C <- serverMsg
	default:
		//向所有用户广播
		u.server.Message <- fmt.Sprintf("「%s」: %s", u.Name, msg)
	}
}

// 向用户发送消息
func (u *User) sendMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
