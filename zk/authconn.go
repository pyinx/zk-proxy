package zk

import (
	"net"

	"github.com/golang/glog"
)

// AuthConn transfers zookeeper handshaking for establishing a session
type AuthConn interface {
	Read() (*AuthRequest, error)
	Write(AuthResponse) (Conn, error)
	WriteFlw(string, string) error
	Close()
}

type AuthResponse struct {
	Resp           *ConnectResponse
	FourLetterWord string
}

type AuthRequest struct {
	Req            *ConnectRequest
	FourLetterWord string
}

type authConn struct {
	c net.Conn
}

func NewAuthConn(c net.Conn) AuthConn { return &authConn{c} }

func (ac *authConn) Read() (*AuthRequest, error) {
	req := &ConnectRequest{}
	flw, err := ReadPacket(ac.c, req)
	if err != nil {
		glog.Errorf("read connection request from %s %v", ac.c.RemoteAddr(), err)
		return nil, err
	}
	return &AuthRequest{req, flw}, nil
}

func (ac *authConn) Write(ar AuthResponse) (Conn, error) {
	if err := WritePacket(ac.c, ar.Resp); err != nil {
		glog.Errorf("response connection request to %s %v", ac.c.RemoteAddr(), err)
		return nil, err
	}
	zkc := NewConn(ac.c)
	ac.c = nil
	return zkc, nil
}

func (ac *authConn) WriteFlw(flw, server string) error {
	return ProxyFlw(ac.c, server, flw)
}

func (ac *authConn) Close() {
	if ac.c != nil {
		ac.c.Close()
	}
}
