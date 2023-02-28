package jrpc

import "context"

type Conn interface {
	Call(ctx context.Context, result any, method string, params ...any) error
	BatchCall(ctx context.Context, b ...BatchElem) error
	SetHeader(key, value string)
	Close() error
}

type SubscriptionConn interface {
	Conn

	Notify(ctx context.Context, method string, args ...any) error
	Subscribe(ctx context.Context, namespace string, channel any, args ...any) (*ClientSubscription, error)
}
