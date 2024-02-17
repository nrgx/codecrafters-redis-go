package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("error creating tcp server", err.Error())
		os.Exit(1)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error establishing connection", err.Error())
			os.Exit(1)
		}
		go handle(conn)
	}
}

func handle(c net.Conn) {
	// defer c.Close()
	c.Write([]byte("+PONG\r\n"))
}
