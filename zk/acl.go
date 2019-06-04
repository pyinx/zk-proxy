package zk

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/samuel/go-zookeeper/zk"
)

type AclCache struct {
	mu sync.RWMutex
	m  map[string]map[string]bool
	t  *time.Ticker
}

var (
	aclCache        = AclCache{t: time.NewTicker(1 * time.Second), m: make(map[string]map[string]bool)}
	zkConn          *zk.Conn
	permRoot        = "/perm"
	enableIPAcl     = false
	errNotEnableAcl = errors.New("not enable ip acl")
)

func InitAcl(servers []string) {
	enableIPAcl = true
	var err error
	zkConn, _, err = zk.Connect(servers, 5*time.Second, zk.WithLogInfo(false))
	if err != nil {
		panic(err)
	}
	exists, _, _ := zkConn.Exists(permRoot)
	if !exists {
		_, err = zkConn.Create(permRoot, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			exists, _, _ = zkConn.Exists(permRoot)
			if !exists {
				panic(err)
			}
		}
	}
	go updateAcl()
}

func updateAcl() {
	for range aclCache.t.C {
		aclCache.mu.RLock()
		secondPath, _, err := zkConn.Children("/")
		if err != nil {
			glog.Errorf("list second path %v", err)
			continue
		}
		for _, p := range secondPath {
			p = "/" + p
			if p == permRoot {
				continue
			}
			pp := permRoot + p
			exists, _, err := zkConn.Exists(pp)
			if err != nil {
				glog.Errorf("exists path %s %v", pp, err)
				continue
			}
			if !exists {
				_, err = zkConn.Create(pp, []byte("{}"), 0, zk.WorldACL(zk.PermAll))
				if err != nil {
					glog.Errorf("create path %s %v", pp, err)
					continue
				} else {
					aclCache.m[p] = make(map[string]bool)
				}
			} else {
				aclMap, _ := getAclData(pp)
				aclCache.m[p] = aclMap
			}
		}
		aclCache.mu.RUnlock()
	}
}

func createLock(path string) error {
	_, err := zkConn.Create(path, []byte{}, 1, zk.WorldACL(zk.PermAll))
	return err
}

func getLock(path string) error {
	lockPath := permRoot + path + "/lock"
	var err error
	timeout := time.NewTimer(1 * time.Second)
	for {
		select {
		case <-timeout.C:
			return errors.New("get lock timeout")
		default:
			err = createLock(lockPath)
			if err != nil {
				time.Sleep(100 * time.Nanosecond)
			} else {
				return nil
			}
		}
	}
}

func unLock(path string) {
	lockPath := permRoot + path + "/lock"
	zkConn.Delete(lockPath, 0)
}

func checkPath(path string) error {
	exists, _, err := zkConn.Exists(path)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(path + " is not exists")
	}
	return nil
}

func getAclData(path string) (aclMap map[string]bool, err error) {
	aclData, _, err := zkConn.Get(path)
	if err != nil {
		glog.Errorf("get acl for path %s %v", path, err)
		return
	}
	err = json.Unmarshal(aclData, &aclMap)
	if err != nil {
		glog.Errorf("load acl for path %s %v", path, err)
	}
	return
}

func saveAclData(path string, aclMap map[string]bool) error {
	aclData, _ := json.Marshal(aclMap)
	_, err := zkConn.Set(path, aclData, -1)
	return err
}

func AddIpAcl(path string, ipaddrs []string) error {
	if !enableIPAcl {
		return errNotEnableAcl
	}
	err := checkPath(path)
	if err != nil {
		return err
	}
	err = getLock(path)
	if err != nil {
		return err
	}
	pp := permRoot + path
	aclMap, _ := getAclData(pp)
	for _, ipaddr := range ipaddrs {
		aclMap[ipaddr] = true
	}
	err = saveAclData(pp, aclMap)
	unLock(path)
	return err
}

func DelIpAcl(path string, ipaddrs []string) error {
	if !enableIPAcl {
		return errNotEnableAcl
	}
	err := checkPath(path)
	if err != nil {
		return err
	}
	err = getLock(path)
	if err != nil {
		return err
	}
	pp := permRoot + path
	aclMap, _ := getAclData(pp)
	for _, ipaddr := range ipaddrs {
		delete(aclMap, ipaddr)
	}
	err = saveAclData(pp, aclMap)
	unLock(path)
	return err
}

func ListIpAcl(path string) (ipList []string, err error) {
	if !enableIPAcl {
		err = errNotEnableAcl
		return
	}
	aclMap, ok := aclCache.m[path]
	if !ok {
		err = errors.New(path + " is not exists")
		return
	}
	for ip := range aclMap {
		ipList = append(ipList, ip)
	}
	return
}

func CheckIpAcl(path string, ipaddr string) bool {
	if !enableIPAcl {
		return true
	}
	if path == "/" || path == "" {
		return true
	}
	secondPath := strings.Join(strings.Split(path, "/")[:2], "/")
	if secondPath == permRoot {
		return false
	}
	aclMap, ok := aclCache.m[secondPath]
	if !ok {
		return true
	} else {
		_, ok = aclMap[ipaddr]
		return ok
	}
	return false
}
