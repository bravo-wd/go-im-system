package main

func main() {
	server := newServer("l127.0.0.1", 8936)
	server.Start()
}
