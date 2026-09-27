package main

import (
	"flag"
	"fmt"
)

var serverIP string
var serverPort int

func init() {
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "设置服务器IP地址")
	flag.IntVar(&serverPort, "port", 8936, "设置服务器端口")
}

func main() {
	flag.Parse()
	client := newClient(serverIP, serverPort)
	fmt.Println(client.Conn)
}
