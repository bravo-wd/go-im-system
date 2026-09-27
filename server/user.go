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
// 约定协议:
// hello, everyone	直接群发
// /who				查看在线用户
// /nick Alice		修改昵称
// /to Alice hello	私聊，消息可以包含空格
// /quit			断开连接
func (u *User) processMessage(msg string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return
	}
	// 群聊
	if !strings.HasPrefix(msg, "/") {
		u.server.Message <- fmt.Sprintf("「%s」: %s", u.Name, msg)
		return
	}

	cmd, args, _ := strings.Cut(msg, " ")
	args = strings.TrimSpace(args)
	switch cmd {
	case "/who":
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
		return
	case "/nick":
		if args == "" || strings.ContainsAny(args, "\t\r\n") {
			u.C <- "「服务器」: 昵称不能为空，也不能包含空格"
			return
		}
		name := args
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
		return
	case "/to":
		name, body, ok := strings.Cut(args, " ")
		body = strings.TrimSpace(body)
		if !ok || name == "" || body == "" {
			u.C <- "「服务器」: 用法：/to 用户名 消息"
			return
		}

		toMsg := body
		var user *User
		var exist bool
		u.server.mapLock.RLock()
		user, exist = u.server.OnlineMap[name]
		u.server.mapLock.RUnlock()
		if !exist {
			u.C <- fmt.Sprintf("「服务器」: 该用户不存在")
			return
		}
		user.C <- fmt.Sprintf("「%s」私聊: %s", u.Name, toMsg)
		return

	case "/quit":
		u.conn.Close()
	default:
		u.C <- "「服务器」: 未知命令; 可用 /who, /nick, /to, /quit"
	}
}

// 向用户发送消息
func (u *User) sendMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
