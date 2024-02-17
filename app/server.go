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
		fmt.Println("=====accepting=====")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error establishing connection", err.Error())
			continue
		}
		buf := make([]byte, 1024)
		fmt.Println("reading")
		if _, err := conn.Read(buf); err != nil {
			fmt.Println("error reading from connection", err.Error())
			continue
		}
		fmt.Println("read")
		buf = []byte("+PONG\r\n")
		fmt.Println("writing")
		if _, err := conn.Write(buf); err != nil {
			fmt.Println("error writing to connection", err.Error())
			continue
		}
		fmt.Println("wrote")
		conn.Close()
		fmt.Println("=====closing=====\n")
	}
}
