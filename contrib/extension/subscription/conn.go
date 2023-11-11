package subscription

import (
	"context"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

type Conn interface {
	Subscribe(ctx context.Context, namespace string, channel any, args any) (ClientSubscription, error)

	jsonrpc.Conn
}

func UpgradeConn(c jsonrpc.Conn, err error) (Conn, error) {
	if err != nil {
		return nil, err
	}
	if val, ok := c.(jsonrpc.Conn); ok {
		engine := NewWrapClient(val)
		val.Mount(engine.Middleware)
		return engine, nil
	}
	return nil, ErrNotificationsUnsupported
}

type ClientSubscription interface {
	Err() <-chan error
	Unsubscribe() error
	String() string
}
