package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {

	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Server Failed to start")
		return
	}
	fmt.Println("Server Started on port 8080")
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Failed to accept client connection")
			return
		}
		go handle_conn(conn)
	}

}
func handle_conn(con net.Conn) {
	defer con.Close()
	raddr := con.RemoteAddr()
	fmt.Printf("New client connected: %v\n", raddr)
	scanner := bufio.NewScanner(con)
	for scanner.Scan() {
		message := scanner.Text()
		switch message {
		case "exit":
			fmt.Printf("Client: %v disconnected\n", raddr)
			return
		case "PING":
			con.Write([]byte("PONG\n"))
		default:
			fmt.Printf("Message received: %v\n", message)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Connection error from %v: %v\n", con.RemoteAddr(), err)
	} else {
		fmt.Printf("Client disconnected cleanly: %v\n", con.RemoteAddr())
	}
}
