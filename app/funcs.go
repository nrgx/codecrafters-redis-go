package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func run() {
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

func respify(value string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value))
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
		if _, err := conn.Write(parse(buf)); err != nil {
			fmt.Println("error writing to connection", err.Error())
			os.Exit(1)
		}
	}
}

func parse(buf []byte) []byte {
	fields := strings.Fields(string(buf))
	fmt.Println(fields)
	var args []string
	for i := range fields[:len(fields)-1] {
		if !(strings.HasPrefix(fields[i], "*") || strings.HasPrefix(fields[i], "$")) {
			args = append(args, fields[i])
		}
	}
	fmt.Println("ARGS", args)
	cmd := args[0]
	args = args[1:]
	switch strings.ToLower(cmd) {
	case PING:
		return R.pong()
	case ECHO:
		return R.echo(args)
	case GET:
		return R.get(args)
	case SET:
		return R.set(args)
	default:
		return NIL
	}
}
