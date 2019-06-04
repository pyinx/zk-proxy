package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

const (
	zkAddr = "127.0.0.1:2181"
	pAddr  = "127.0.0.1:2182"
)

var acl = zk.WorldACL(zk.PermAll)

func benchGet(b *testing.B, addr string) {
	c, _, err := zk.Connect([]string{addr}, time.Second, zk.WithLogInfo(false))
	if err != nil {
		b.Fatal(err)
	}
	defer c.Close()
	c.Create("/"+addr, []byte(addr), 0, acl)
	for i := 0; i < b.N; i++ {
		if _, _, gerr := c.Get("/" + addr); gerr != nil {
			b.Fatal(err)
		}
	}
}

func benchConnGet(b *testing.B, addr string) {
	for i := 0; i < b.N; i++ {
		c, _, err := zk.Connect([]string{addr}, time.Second, zk.WithLogInfo(false))
		if err != nil {
			b.Fatal(err)
		}
		if _, _, gerr := c.Get("/" + addr); gerr != nil {
			b.Fatal(err)
		}
		c.Close()
	}
}

func benchCreateSet(b *testing.B, addr string) {
	c, _, err := zk.Connect([]string{addr}, time.Second, zk.WithLogInfo(false))
	if err != nil {
		b.Fatal(err)
	}
	defer c.Close()
	for i := 0; i < b.N; i++ {
		s := fmt.Sprintf("/%s/%d", addr, i)
		v := fmt.Sprintf("%v", time.Now())
		c.Create(s, []byte(v), 0, acl)
		c.Set("/", []byte(v), -1)
	}
}

func BenchmarkProxyGet(b *testing.B) { benchGet(b, pAddr) }
func BenchmarkZKGet(b *testing.B)    { benchGet(b, zkAddr) }

func BenchmarkProxyConnGet(b *testing.B) { benchConnGet(b, pAddr) }
func BenchmarkZKConnGet(b *testing.B)    { benchConnGet(b, zkAddr) }

func BenchmarkProxyCreateSet(b *testing.B) { benchCreateSet(b, pAddr) }
func BenchmarkZKCreateSet(b *testing.B)    { benchCreateSet(b, zkAddr) }
