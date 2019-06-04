package zk

import (
	"io"
	"net"
	"runtime"
	"strconv"
	"time"

	"github.com/orcaman/concurrent-map"
)

var activeSessions cmap.ConcurrentMap

func init() {
	activeSessions = cmap.New()
}

func getSess() (response string) {
	if activeSessions.IsEmpty() {
		return ""
	}
	for _, sid := range activeSessions.Keys() {
		s, ok := activeSessions.Get(sid)
		if ok {
			response = response + " /" + s.(Session).ClientAddress() + " " + sid + "\n"
		}
	}
	return
}

func getInfo() (response string) {
	response = response + "num_alive_connections\t" + strconv.Itoa(activeSessions.Count()) + "\n"
	response = response + "go_num_goroutine\t" + strconv.Itoa(runtime.NumGoroutine()) + "\n"
	response = response + "go_num_cgo_call\t" + strconv.Itoa(int(runtime.NumCgoCall())) + "\n"
	response = response + "go_version\t" + runtime.Version() + "\n"
	return
}

func ProxyFlw(rc net.Conn, server, flw string) error {
	if flw == "isok" {
		_, err := rc.Write([]byte("imok\n"))
		return err
	} else if flw == "info" {
		_, err := rc.Write([]byte(getInfo()))
		return err
	} else if flw == "sess" {
		_, err := rc.Write([]byte(getSess()))
		return err
	} else {
		lc, err := net.DialTimeout("tcp", server, time.Duration(500*time.Millisecond))
		if err != nil {
			return err
		}
		defer lc.Close()

		buff := make([]byte, 0xffff)
		_, err = lc.Write([]byte(flw))
		if err != nil {
			return err
		}
		for {
			n, err := lc.Read(buff)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
			b := buff[:n]
			_, err = rc.Write(b)
			if err != nil {
				return err
			}
		}
	}
}
