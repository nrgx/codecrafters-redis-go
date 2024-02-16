package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	c, err := l.Accept()

	for {
		if _, err := c.Write([]byte("+PONG\r\n")); err != nil {
			break
		}
	}
}
