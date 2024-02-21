package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Value struct {
	value  string
	expiry time.Time
}

func newValue(args []string) Value {
	value := args[1]
	res := Value{value: value}
	if len(args) == 4 {
		xKey := strings.ToLower(args[2])
		expiry, err := strconv.Atoi(args[3])
		if err != nil {
			return res
		}
		var ttl time.Duration
		switch xKey {
		case "px":
			ttl = time.Duration(expiry) * time.Millisecond
		case "ex":
			ttl = time.Duration(expiry) * time.Second
		}
		res.expiry = time.Now().Add(ttl)
		fmt.Println("TTL", ttl)
	}
	return res
}

func (v Value) isExpired() bool {
	return time.Now().After(v.expiry)
}

func (v Value) String() string {
	return v.value
}

type REDIS struct {
	data map[string]Value
	mx   *sync.Mutex
}

func New() *REDIS {
	redis := &REDIS{
		data: make(map[string]Value),
		mx:   new(sync.Mutex),
	}

	go func() {
		for range time.Tick(50 * time.Millisecond) {
			redis.checkExpiry()
		}
	}()

	return redis
}

func (r *REDIS) checkExpiry() {
	r.mx.Lock()
	defer r.mx.Unlock()
	for k, v := range r.data {
		if v.isExpired() {
			delete(r.data, k)
		}
	}
}

func (r *REDIS) get(args []string) []byte {
	r.mx.Lock()
	defer r.mx.Unlock()
	k := args[0]
	value, ok := r.data[k]
	if !ok {
		if len(args) == 2 {
			return respify(args[1])
		}
		return NIL
	}
	if value.isExpired() {
		return NIL
	}
	return respify(value.value)
}

func (r *REDIS) set(args []string) []byte {
	if len(args) <= 1 {
		fmt.Println("given args less than required", args)
		return NIL
	}
	r.mx.Lock()
	defer r.mx.Unlock()
	r.data[args[0]] = newValue(args)
	return OK
}

func (r *REDIS) pong() []byte {
	return []byte("+PONG\r\n")
}

func (r *REDIS) echo(args []string) []byte {
	buf := bytes.NewBuffer(nil)
	for _, arg := range args[1:] {
		// write $LEN\r\nARG\r\n
		buf.Write(respify(arg))
	}
	return buf.Bytes()
}

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

var R *REDIS = nil

func init() {
	R = New()
}

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
		if _, err := conn.Write(parse(buf)); err != nil {
			fmt.Println("error writing to connection", err.Error())
			os.Exit(1)
		}
	}
}

func parse(buf []byte) []byte {
	args := strings.Fields(string(buf))
	args = slices.DeleteFunc[[]string, string](args, func(s string) bool {
		return strings.HasPrefix(s, "*") || strings.HasPrefix(s, "$")
	})
	cmd := args[0]
	args = args[1 : len(args)-1]
	fmt.Printf("args:{%s}", args)
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

func respify(value string) []byte {
	fmt.Println("value before respify", value)
	s := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
	fmt.Println("before byte conversion", s)
	out := []byte(s)
	fmt.Printf("respify %s\n", out)
	return out
}
