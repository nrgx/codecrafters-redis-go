package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

var pong = []byte("+PONG\r\n")

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("error creating tcp server", err.Error())
		os.Exit(1)
	}
	defer listener.Close()
	for {
		fmt.Println("=====accepting=====")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error establishing connection", err.Error())
			continue
		}
		fmt.Println("connection established")
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		fmt.Println("reading")
		if _, err := conn.Read(buf); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error reading from connection", err.Error())
			os.Exit(1)
		}
		fmt.Println("read")
		fmt.Println("writing")
		if _, err := conn.Write(pong); err != nil {
			fmt.Println("error writing to connection", err.Error())
			os.Exit(1)
		}
		fmt.Println("wrote")
	}

	conn.Close()
	fmt.Println("======closing======\n")
}
