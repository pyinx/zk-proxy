package zk

import (
	"encoding/binary"
	"io"
	"net"
)

// client represents a client that connects to a zk server.
type client struct {
	zkc   net.Conn
	readc chan ZKResponse
	stopc chan struct{}
}

type Client interface {
	Send(resp []byte) (int, error)
	Read() <-chan ZKResponse
	Close()
	RemoteAddress() string
}

type ZKResponse struct {
	hdr *ResponseHeader
	err error
	raw []byte
}

func NewClient(zk net.Conn) Client {
	c := &client{
		zkc:   zk,
		readc: make(chan ZKResponse),
		stopc: make(chan struct{}),
	}

	go func() {
		defer close(c.readc)
		for {
			buf, hdr, err := readRespOp(c.zkc)
			select {
			case c.readc <- ZKResponse{hdr: hdr, err: err, raw: buf}:
				if err != nil && err != io.EOF {
					return
				}
			case <-c.stopc:
				return
			}
		}
	}()
	return c
}

func (c *client) Read() <-chan ZKResponse { return c.readc }

func (c *client) Send(resp []byte) (int, error) {
	buf := make([]byte, 4)
	buf = append(buf, resp...)
	binary.BigEndian.PutUint32(buf[:4], uint32(len(resp)))
	return c.zkc.Write(buf)
}

func (c *client) Close() {
	close(c.stopc)
	c.zkc.Close()
}

func (c *client) RemoteAddress() string { return c.zkc.RemoteAddr().String() }
