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
			continue
		}
		go handle(conn)
	}
}

func handle(c net.Conn) {
	defer c.Close()
	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	if err != nil {
		fmt.Println("error reading from connection", err.Error())
		os.Exit(1)
	}
	fmt.Print(string(buf[8:n]))
	if _, err := c.Write([]byte("+PONG\r\n")); err != nil {
		fmt.Println("error writing to connection", err.Error())
		os.Exit(1)
	}
}
