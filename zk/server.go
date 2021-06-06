package zk

import (
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/uber-go/atomic"
	"golang.org/x/net/context"
)

var (
	isLimit  = false
	limitNum = 1000 // per second
)

type acceptHandler func(ctx context.Context, conn net.Conn, auth AuthFunc, zk ZKFunc)

func SetLimit(num int) {
	isLimit = true
	limitNum = num
	glog.V(1).Infof("set request rate limit for %d", limitNum)
}

func Serve(ctx context.Context, ln net.Listener, auth AuthFunc, zk ZKFunc) {
	if isLimit {
		serveByHandler(ctx, handleSessionConcurrentRequests, ln, auth, zk)
	} else {
		serveByHandler(ctx, handleSessionSerialRequests, ln, auth, zk)
	}
}

func handleSessionSerialRequests(ctx context.Context, conn net.Conn, auth AuthFunc, zk ZKFunc) {
	s, zke, serr := openClientSession(ctx, conn, auth, zk)
	if serr != nil {
		return
	}
	glog.V(1).Infof("serving serial session requests from %s session %s", conn.RemoteAddr(), s.SidStr())
	for zkreq := range s.Read() {
		if err := serveRequest(s, zke, zkreq); err != nil {
			s.SClose()
			return
		}
	}
}

func handleSessionConcurrentRequests(ctx context.Context, conn net.Conn, auth AuthFunc, zk ZKFunc) {
	rl := ratelimit.New(limitNum)

	s, zke, serr := openClientSession(ctx, conn, auth, zk)
	if serr != nil {
		return
	}
	glog.V(1).Infof("serving concurrent session requests from %s session %s", conn.RemoteAddr(), s.SidStr())
	for zkreq := range s.Read() {
		rl.Take()
		if err := serveRequest(s, zke, zkreq); err != nil {
			s.SClose()
			return
		}
	}
}

// receive request from client and send respone to client
func serveRequest(s Session, zke ZK, zkreq ZKRequest) error {
	if zkreq.err != nil {
		return zkreq.err
	}
	st := time.Now()
	opType, zkPath, respErr := DispatchZK(zke, zkreq.xid, zkreq.req, zkreq.raw)
	if respErr != nil {
		glog.Errorf("dispatch %+v for %s err %s", zkreq, s.SidStr(), respErr)
	}
	glog.V(0).Infof("%s %s %s %s %s %.6f", s.ClientAddress(), s.ServerAddress(), s.SidStr(), opType, zkPath, time.Since(st).Seconds())
	return respErr
}

func openClientSession(ctx context.Context, conn net.Conn, auth AuthFunc, zk ZKFunc) (Session, ZK, error) {
	glog.V(1).Infof("accepted client connection from %s", conn.RemoteAddr())
	s, serr := auth(ctx, NewAuthConn(conn))
	if serr != nil {
		return nil, nil, serr
	}
	glog.V(1).Infof("open new session for client %s %s", conn.RemoteAddr(), formatZkId(int64(s.Sid())))
	return s, zk(s), nil
}

func serveByHandler(ctx context.Context, h acceptHandler, ln net.Listener, auth AuthFunc, zk ZKFunc) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			glog.Errorf("Accept err %v", err)
			return
		} else {
			go h(ctx, conn, auth, zk)
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}
