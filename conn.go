package jrpc

import "context"

type Conn interface {
	Do(ctx context.Context, result any, method string, params any) error
	BatchCall(b ...BatchElem) error
}

type SubscriptionConn interface {
	Notify(ctx context.Context, method string, args ...any) error
	Subscribe(ctx context.Context, namespace string, channel any, args ...any) (*ClientSubscription, error)
}
