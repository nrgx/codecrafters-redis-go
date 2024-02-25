package main

var R *REDIS

func init() {
	R = New()
}

func main() {
	run()
}
