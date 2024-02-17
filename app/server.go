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

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("error establishing connection", err.Error())
		os.Exit(1)
	}
	for {
		go func(c net.Conn) {
			if _, err := c.Write([]byte("+PONG\r\n")); err != nil {
				fmt.Println("error writing to connection", err.Error())
				os.Exit(1)
			}
		}(conn)
	}
	c.Close()
}
