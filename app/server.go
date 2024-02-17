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

	defer func() {
		if err := listener.Close(); err != nil {
			fmt.Println("error closing tcp server", err.Error())
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error establishing connection", err.Error())
			os.Exit(1)
		}
		go func(c net.Conn) {
			if _, err := c.Write([]byte("+PONG\r\n")); err != nil {
				fmt.Println("error writing to connection", err.Error())
				os.Exit(1)
			}
			c.Close()
		}(conn)
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
	fmt.Println(buf[8:n])
	if _, err := c.Write([]byte("+PONG\r\n")); err != nil {
		fmt.Println("error writing to connection", err.Error())
		os.Exit(1)
	}
}
