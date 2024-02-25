package main

import (
	"bytes"
	"fmt"
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
		for range time.Tick(100 * time.Millisecond) {
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
	for _, arg := range args {
		// write $LEN\r\nARG\r\n
		buf.Write(respify(arg))
	}
	return buf.Bytes()
}
