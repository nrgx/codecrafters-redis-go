package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection", err.Error())
			os.Exit(1)
		}
		buf := make([]byte, 1024)
		if _, err := conn.Read(buf); err != nil {
			fmt.Println("Error reading from connection", err.Error())
			os.Exit(1)
		}
		buf = []byte("+PONG\r\n")
		if _, err := conn.Write(buf); err != nil {
			fmt.Println("Error writing to connection", err.Error())
			os.Exit(1)
		}
	}
}
