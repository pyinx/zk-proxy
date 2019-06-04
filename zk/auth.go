package zk

import (
	"golang.org/x/net/context"
)

type AuthFunc func(context.Context, AuthConn) (Session, error)

type ZKFunc func(Session) ZK

func NewAuth(servers []string) AuthFunc {
	return func(ctx context.Context, zka AuthConn) (Session, error) {
		return newSession(ctx, servers, zka)
	}
}

func NewZK() ZKFunc {
	return func(s Session) ZK {
		return newZK(s)
	}
}
