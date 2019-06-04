package zk

import (
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/context"
)

type Session interface {
	Conn
	Sid() Sid
	SidStr() string
	ConnReq() ConnectRequest
	ClientAddress() string
	ServerAddress() string
	SClose()
}

type session struct {
	Conn
	zkc     Client
	connReq ConnectRequest
	sid     Sid
	sidStr  string

	ctx    context.Context
	cancel context.CancelFunc

	clientAddress string
	serverAddress string
}

func (s *session) Sid() Sid                { return s.sid }
func (s *session) SidStr() string          { return s.sidStr }
func (s *session) ClientAddress() string   { return s.clientAddress }
func (s *session) ServerAddress() string   { return s.serverAddress }
func (s *session) ConnReq() ConnectRequest { return s.connReq }

func (s *session) SClose() { s.cancel() }

func (s *session) close() {
	activeSessions.Remove(s.SidStr())
	s.Conn.Close()
	s.zkc.Close()
}

func newSession(ctx context.Context, servers []string, zka AuthConn) (*session, error) {
	defer zka.Close()
	// read request from client
	areq, err := zka.Read()
	if err != nil {
		return nil, err
	}
	// if is flw, return response
	if areq.FourLetterWord != "" {
		shuffleZkServer(servers)
		aerr := zka.WriteFlw(areq.FourLetterWord, servers[0])
		if aerr != nil {
			glog.Errorf("failed to proxy fourLetterWord %s %v", areq.FourLetterWord, aerr)
		} else {
			aerr = ErrFourLetterWord
		}
		return nil, aerr
	}

	resp := ConnectResponse{}
	zkConn, err := dialZKServer(servers)
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}
	// send connection request
	if err = WritePacket(zkConn, areq.Req); err != nil {
		glog.Errorf("failed to write connection request to %s %v", zkConn.RemoteAddr(), err)
		zkConn.Close()
		return nil, err
	}
	// pipe back connection result
	flw, err := ReadPacket(zkConn, &resp)
	if err != nil {
		glog.Errorf("failed to read connection response from %s %v", zkConn.RemoteAddr(), err)
		zkConn.Close()
		return nil, err
	}
	// proxy response to client
	zkc, aerr := zka.Write(AuthResponse{Resp: &resp, FourLetterWord: flw})
	if zkc == nil || aerr != nil {
		glog.Errorf("failed to auth from %s %v", zkConn.RemoteAddr(), err)
		zkConn.Close()
		return nil, aerr
	}

	sessionCtx, cancel := context.WithCancel(ctx)
	req := ConnectRequest{}
	s := &session{
		Conn:    zkc,
		zkc:     NewClient(zkConn),
		connReq: req,
		sid:     resp.SessionID,
		sidStr:  formatZkId(int64(resp.SessionID)),
		ctx:     sessionCtx,
		cancel:  cancel,
	}
	s.clientAddress = s.Conn.RemoteAddress()
	s.serverAddress = s.zkc.RemoteAddress()

	activeSessions.Set(s.SidStr(), s)
	go s.recvLoop()
	return s, nil
}

func (s *session) future(xid Xid, path string, raw []byte) error {
	clientAddr := strings.Split(s.clientAddress, ":")[0]
	if !CheckIpAcl(path, clientAddr) {
		glog.Warningf("auth failed: client addr: %s path: %s", clientAddr, path)
		raw, _ = generateErrResp(xid, errNoAuth)
		_, err := s.Send(raw)
		if err != nil {
			glog.Errorf("send acl err to client for %d %v", int(xid), err)
		}
		return err
	}
	_, err := s.zkc.Send(raw)
	if err != nil {
		glog.Errorf("send request to zk server for %d %v", int(xid), err)
	}
	return err
}

// recvLoop forwards responses from the real zk server to the client connection.
func (s *session) recvLoop() {
	defer s.close()
	for {
		select {
		case resp := <-s.zkc.Read():
			if resp.err != nil {
				if resp.err != io.EOF {
					glog.Errorf("receloop read data from zk server %v", resp.err)
				}
				return
			}
			_, err := s.Send(resp.raw)
			if err != nil {
				glog.Errorf("receloop send data to client %v", err)
				return
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func GetZkServers(servers string) []string {
	serverList := strings.Split(servers, ",")
	srvs := make([]string, len(serverList))
	for i, addr := range serverList {
		if strings.Contains(addr, ":") {
			srvs[i] = addr
		} else {
			srvs[i] = addr + ":" + strconv.Itoa(DefaultPort)
		}
	}
	return srvs
}

func shuffleZkServer(servers []string) {
	for i := len(servers) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		servers[i], servers[j] = servers[j], servers[i]
	}
}

func dialZKServer(servers []string) (net.Conn, error) {
	var conn net.Conn
	var err error
	shuffleZkServer(servers)
	for _, zkServer := range servers {
		conn, err = net.DialTimeout("tcp", zkServer, time.Duration(500*time.Millisecond))
		if err == nil {
			return conn, err
		}
		glog.V(3).Infof("connect %s err %s", zkServer, err)
	}
	return nil, ErrNoServer
}

func formatZkId(id int64) string {
	return "0x" + strconv.FormatInt(id, 16)
}
