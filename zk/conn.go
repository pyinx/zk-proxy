package zk

import (
	"encoding/binary"
	"fmt"
	"net"
)

type Conn interface {
	Send(resp []byte) (int, error)
	Read() <-chan ZKRequest
	Close()
	RemoteAddress() string
}

type conn struct {
	zkc   net.Conn
	readc chan ZKRequest
	stopc chan struct{}
}

type ZKRequest struct {
	xid Xid
	req interface{}
	err error
	raw []byte
}

func (zk *ZKRequest) String() string {
	if zk.req != nil {
		return fmt.Sprintf("{xid:%v req:%T:%+v}", zk.xid, zk.req, zk.req)
	}
	if zk.err != nil {
		return fmt.Sprintf("{xid:%v err:%q}", zk.xid, zk.err)
	}
	return fmt.Sprintf("{xid:%v err:%q}", zk.xid, zk.err)
}

func NewConn(zk net.Conn) Conn {
	c := &conn{
		zkc:   zk,
		readc: make(chan ZKRequest),
		stopc: make(chan struct{}),
	}

	go func() {
		defer close(c.readc)
		for {
			buf, xid, req, err := readReqOp(c.zkc)
			select {
			case c.readc <- ZKRequest{xid, req, err, buf}:
				if err != nil {
					return
				}
			case <-c.stopc:
				return
			}
		}
	}()

	return c
}

func (c *conn) Read() <-chan ZKRequest { return c.readc }

func (c *conn) Send(resp []byte) (int, error) {
	buf := make([]byte, 4)
	buf = append(buf, resp...)
	binary.BigEndian.PutUint32(buf[:4], uint32(len(resp)))
	return c.zkc.Write(buf)
}

func (c *conn) Close() {
	close(c.stopc)
	c.zkc.Close()
}

func (c *conn) RemoteAddress() string { return c.zkc.RemoteAddr().String() }
