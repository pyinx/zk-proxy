package zk

import (
	"github.com/golang/glog"
)

type ZK interface {
	Create(xid Xid, path string, raw []byte) error
	Delete(xid Xid, path string, raw []byte) error
	Exists(xid Xid, path string, raw []byte) error
	GetData(xid Xid, path string, raw []byte) error
	SetData(xid Xid, path string, raw []byte) error
	GetAcl(xid Xid, path string, raw []byte) error
	SetAcl(xid Xid, path string, raw []byte) error
	GetChildren(xid Xid, path string, raw []byte) error
	Sync(xid Xid, path string, raw []byte) error
	Ping(xid Xid, path string, raw []byte) error
	GetChildren2(xid Xid, path string, raw []byte) error
	Multi(xid Xid, path string, raw []byte) error
	Close(xid Xid, path string, raw []byte) error
	SetAuth(xid Xid, path string, raw []byte) error
	SetWatches(xid Xid, path string, raw []byte) error
}

type zkZK struct{ s *session }

func newZK(s Session) ZK {
	return &zkZK{s.(*session)}
}

func (zz *zkZK) Create(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) Delete(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) Exists(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) GetData(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) SetData(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) GetAcl(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) SetAcl(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) GetChildren(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) Sync(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) Ping(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) GetChildren2(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) Multi(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) Close(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) SetAuth(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}
func (zz *zkZK) SetWatches(xid Xid, path string, raw []byte) error {
	return zz.s.future(xid, path, raw)
}

func DispatchZK(zk ZK, xid Xid, op interface{}, raw []byte) (string, string, error) {
	switch op := op.(type) {
	case *CreateRequest:
		return "Create", op.Path, zk.Create(xid, op.Path, raw)
	case *DeleteRequest:
		return "Delete", op.Path, zk.Delete(xid, op.Path, raw)
	case *GetChildrenRequest:
		return "GetChildren", op.Path, zk.GetChildren(xid, op.Path, raw)
	case *GetChildren2Request:
		return "GetChildren2", op.Path, zk.GetChildren2(xid, op.Path, raw)
	case *PingRequest:
		return "Ping", "", zk.Ping(xid, "", raw)
	case *GetDataRequest:
		return "Get", op.Path, zk.GetData(xid, op.Path, raw)
	case *SetDataRequest:
		return "Set", op.Path, zk.SetData(xid, op.Path, raw)
	case *ExistsRequest:
		return "Exists", op.Path, zk.Exists(xid, op.Path, raw)
	case *SyncRequest:
		return "Sync", op.Path, zk.Sync(xid, op.Path, raw)
	case *CloseRequest:
		return "Close", "", zk.Close(xid, "", raw)
	case *SetWatchesRequest:
		return "SetWatches", "", zk.SetWatches(xid, "", raw)
	case *MultiRequest:
		return "Multi", "", zk.Multi(xid, "", raw)
	case *GetAclRequest:
		return "GetAcl", op.Path, zk.GetAcl(xid, op.Path, raw)
	case *SetAclRequest:
		return "SetAcl", op.Path, zk.SetAcl(xid, op.Path, raw)
	case *SetAuthRequest:
		return "SetAuth", "", zk.SetAuth(xid, "", raw)
	default:
		glog.Errorf("unexpected type %d %T\n", xid, op, raw)
	}
	return "Unknown", "", ErrAPIError
}
