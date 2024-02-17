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
	for {
		// this blocks loop
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error accepting connection", err.Error())
			os.Exit(1)
		}
		buf := make([]byte, 1024)
		if _, err := conn.Read(buf); err != nil {
			fmt.Println("error reading from connection", err.Error())
			os.Exit(1)
		}
		// first 8 bytes are header
		// fmt.Println(string(buf[8:n]))
		buf = []byte("+PONG\r\n")
		if _, err := conn.Write(buf); err != nil {
			fmt.Println("error writing to connection", err.Error())
			os.Exit(1)
		}
		conn.Close()
	}
}

