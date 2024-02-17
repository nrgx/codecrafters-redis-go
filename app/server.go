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
		fmt.Println("running")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error accepting connection", err.Error())
			os.Exit(1)
		}
		fmt.Println("accpeted")
		go func(c net.Conn) {
			fmt.Println("reading")
			buf := make([]byte, 1024)
			n, err := c.Read(buf)
			if err != nil {
				fmt.Println("error reading from connection", err.Error())
				os.Exit(1)
			}
			// first 8 bytes are header
			fmt.Println(string(buf[8:n]))
			buf = []byte("+PONG\r\n")
			fmt.Println("writing")
			if _, err := c.Write(buf); err != nil {
				fmt.Println("error writing to connection", err.Error())
				os.Exit(1)
			}
			c.Close()
		}(conn)
		fmt.Println("finished")
	}
}
