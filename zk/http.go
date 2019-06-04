package zk

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"
)

type respBody struct {
	Code int    `json:"code"`
	Err  string `json:"err"`
	Msg  string `json:"msg"`
}

func StartHttp(apiAddr string) {
	http.HandleFunc("/api/v1/whitelist/add", AddIpWhitelist)
	http.HandleFunc("/api/v1/whitelist/del", DelIpWhitelist)
	http.HandleFunc("/api/v1/whitelist/list", ListIpWhitelist)
	srv := &http.Server{
		Addr:         apiAddr,
		WriteTimeout: 3 * time.Second,
		ReadTimeout:  3 * time.Second,
	}
	go srv.ListenAndServe()
}

func GetMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}

func AddIpWhitelist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp := respBody{}
	r.ParseForm()
	pathArg, found1 := r.Form["path"]
	iplistArg, found2 := r.Form["iplist"]

	if !(found1 && found2) {
		resp.Code = -1
		resp.Err = "invalid input args"
		fmt.Fprint(w, marshalResp(resp))
		return
	}
	path := pathArg[0]
	iplist := iplistArg[0]
	if !strings.HasPrefix(path, "/") || strings.Count(path, "/") != 1 {
		resp.Code = -1
		resp.Err = "invalid input args: path"
		fmt.Fprint(w, marshalResp(resp))
		return
	}
	ipaddrs := []string{}
	for _, ip := range strings.Split(iplist, ",") {
		if net.ParseIP(ip) == nil {
			resp.Code = -1
			resp.Err = "invalid input args: " + ip
			fmt.Fprint(w, marshalResp(resp))
			return
		} else {
			ipaddrs = append(ipaddrs, ip)
		}
	}
	err := AddIpAcl(path, ipaddrs)
	glog.V(1).Infof("[Client:%s] [URI:%s] [Path:%s] [IPList:%s] [Err:%v]", r.RemoteAddr, r.RequestURI, path, iplist, err)
	if err != nil {
		resp.Code = 1
		resp.Err = err.Error()
		fmt.Fprint(w, marshalResp(resp))
		return
	}
	resp.Code = 0
	resp.Msg = "success"
	fmt.Fprint(w, marshalResp(resp))
}

func DelIpWhitelist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp := respBody{}
	r.ParseForm()
	pathArg, found1 := r.Form["path"]
	iplistArg, found2 := r.Form["iplist"]

	if !(found1 && found2) {
		resp.Code = -1
		resp.Err = "invalid input args"
		fmt.Fprint(w, marshalResp(resp))
		return
	}
	path := pathArg[0]
	iplist := iplistArg[0]
	if !strings.HasPrefix(path, "/") || strings.Count(path, "/") != 1 {
		resp.Code = -1
		resp.Err = "invalid input args: path"
		fmt.Fprint(w, marshalResp(resp))
		return
	}
	ipaddrs := []string{}
	for _, ip := range strings.Split(iplist, ",") {
		if net.ParseIP(ip) == nil {
			resp.Code = -1
			resp.Err = "invalid input args: " + ip
			fmt.Fprint(w, marshalResp(resp))
			return
		} else {
			ipaddrs = append(ipaddrs, ip)
		}
	}
	err := DelIpAcl(path, ipaddrs)
	glog.V(1).Infof("[Client:%s] [URI:%s] [Path:%s] [IPList:%s] [Err:%v]", r.RemoteAddr, r.RequestURI, path, iplist, err)
	if err != nil {
		resp.Code = 1
		resp.Err = err.Error()
		fmt.Fprint(w, marshalResp(resp))
		return
	}
	resp.Code = 0
	resp.Msg = "success"
	fmt.Fprint(w, marshalResp(resp))
}

func ListIpWhitelist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp := respBody{}
	pathArg, ok := r.URL.Query()["path"]

	if !ok {
		resp.Code = -1
		resp.Err = "invalid input args"
		fmt.Fprint(w, marshalResp(resp))
		return
	}
	path := pathArg[0]
	ipList, err := ListIpAcl(path)
	if err != nil {
		resp.Code = 1
		resp.Err = err.Error()
		fmt.Fprint(w, marshalResp(resp))
		return
	}
	resp.Code = 0
	resp.Msg = strings.Join(ipList, ",")
	fmt.Fprint(w, marshalResp(resp))
}

func marshalResp(r respBody) string {
	b, _ := json.Marshal(r)
	return string(b)
}
