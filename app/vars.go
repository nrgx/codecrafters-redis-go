package main

const (
	PING = "ping"
	ECHO = "echo"
	GET  = "get"
	SET  = "set"
)

var (
	NIL = []byte("$-1\r\n")
	OK  = []byte("+OK\r\n")
)
