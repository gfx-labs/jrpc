package jsonrpc

import (
	"context"
	"io"
)

type Conn interface {
	Doer
	Notifier
	BatchCaller

	Mounter

	io.Closer
	Closed() <-chan struct{}
}

type Doer interface {
	Do(ctx context.Context, result any, method string, params any) error
}

type BatchCaller interface {
	BatchCall(ctx context.Context, b ...*BatchElem) error
}

type Notifier interface {
	Notify(ctx context.Context, method string, params any) error
}

type Mounter interface {
	Mount(Middleware)
}
