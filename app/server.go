package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("error creating tcp server", err.Error())
		os.Exit(1)
	}
	defer listener.Close()
	for {
		// accept blocks loop
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error establishing connection", err.Error())
			continue
		}
		// spawn goroutine in background and pass it a connection
		go process(conn)
	}
}

func process(conn net.Conn) {
	// till connection is live read and write inside a loop
	// if EOF found break
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		if _, err := conn.Read(buf); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error reading from connection", err.Error())
			os.Exit(1)
		}
		out := parse(buf)
		fmt.Printf("out: %q\n", out)
		if _, err := conn.Write(out); err != nil {
			fmt.Println("error writing to connection", err.Error())
			os.Exit(1)
		}
	}
} 

func parse(buf []byte) []byte {
	if len(buf) < 3 {
		return nil
	}
	var start int
	var commands [][]byte
	for i := 0; i < len(buf); i++ {
		if buf[i] == '$' {
			length := int(buf[i+1] - 48)
			start = i + 4
			commands = append(commands, buf[start:start+length])
		}
	}
	fmt.Printf("commands: %q\n", commands)
	switch strings.ToLower(string(commands[0])) {
	case "ping":
		return pong()
	case "echo":
		return echo(commands[1:])
	}
	return nil
}

func pong() []byte {
	return []byte("+PONG\r\n")
}

func echo(args [][]byte) []byte {
	buf := bytes.NewBuffer(nil)
	for _, arg := range args {
		// write $LEN\r\nARG\r\n
		buf.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), string(arg))))
	}
	return buf.Bytes()
}
