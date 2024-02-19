package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Value[V any] struct {
	value  V
	expiry time.Time
}

func (v Value[V]) Value() string {
	return fmt.Sprintf("%s", v.value)
}
func (v Value[V]) String() string {
	return fmt.Sprintf("value:{%v} | expiry: {%v}", v.value, time.Since(v.expiry))
}

func newVal[V any](value V, expKey string, expiry int) Value[V] {
	v := Value[V]{}
	v.value = value
	key := strings.ToLower(expKey)
	var ttl time.Duration
	if key == "px" {
		ttl = time.Duration(expiry) * time.Millisecond
	} else if key == "ex" {
		ttl = time.Duration(expiry) * time.Second
	}
	fmt.Println("TTL", ttl)
	v.expiry = time.Now().Add(ttl)
	return v
}

func (v Value[V]) isExpired() bool {
	return time.Now().After(v.expiry)
}

type REDIS[K comparable, V any] struct {
	data map[K]Value[V]
	mx   *sync.Mutex
}

func New[K comparable, V any]() *REDIS[K, V] {
	redis := &REDIS[K, V]{
		data: make(map[K]Value[V]),
		mx:   new(sync.Mutex),
	}

	go func() {
		for {
			redis.checkExpiry()
		}
	}()

	return redis
}

func (r *REDIS[K, V]) get(k K, val string) []byte {
	r.mx.Lock()
	defer r.mx.Unlock()
	fmt.Println("key:", k, "check:", r.data[k])
	v, ok := r.data[k]
	if !ok {
		fmt.Println("not ok", r.data[k])
		if val != "" {
			return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
		}
		return NIL
	}
	length, err := getLen(v)
	if err != nil {
		fmt.Println("error length", err)
		return NIL
	}
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", length, v.Value()))
}

func (r *REDIS[K, V]) checkExpiry() {
	r.mx.Lock()
	defer r.mx.Unlock()
	for k, v := range r.data {
		if v.isExpired() {
			fmt.Println("deleting")
			delete(r.data, k)
		}
	}
}

func getLen(v interface{}) (int, error) {
	switch val := v.(type) {
	case string:
		return len(val), nil
	case []interface{}:
		return len(val), nil
	case map[interface{}][]interface{}:
		return len(val), nil
	case Value[string]:
		return len(val.value), nil
	default:
		return 0, fmt.Errorf("value `%v` of unknown type `%T`", val, val)
	}
}

func (r *REDIS[K, V]) set(k K, v Value[V]) []byte {
	r.mx.Lock()
	defer r.mx.Unlock()
	r.data[k] = v
	fmt.Println(r.data)
	return OK
}

func (r *REDIS[K, V]) pong() []byte {
	return []byte("+PONG\r\n")
}

func (r *REDIS[K, V]) echo(args [][]byte) []byte {
	buf := bytes.NewBuffer(nil)
	for _, arg := range args {
		// write $LEN\r\nARG\r\n
		buf.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), string(arg))))
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
	// null bulk string RESP encoded
	NIL = []byte("$-1\r\n")
	OK  = []byte("+OK\r\n")
)

var R *REDIS[string, string] = nil

func init() {
	R = New[string, string]()
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
	case PING:
		return R.pong()
	case ECHO:
		return R.echo(commands[1:])
	case GET:
		key := string(commands[1])
		if len(commands) > 2 {
			tmp := string(commands[2])
			return R.get(key, tmp)
		}
		return R.get(key, "")
	case SET:
		if len(commands) == 5 {
			ttl, err := strconv.Atoi(string(commands[4]))
			if err != nil {
				return NIL
			}
			v := newVal[string](string(commands[2]), string(commands[3]), ttl)
			return R.set(string(commands[1]), v)
		}
		return R.set(string(commands[1]), Value[string]{value: string(commands[2])})
	}
	return nil
}
