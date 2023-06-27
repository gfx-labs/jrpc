package codec

import (
	"context"
	"io"
)

type Doer interface {
	Do(ctx context.Context, result any, method string, params any) error
}

type BatchCaller interface {
	BatchCall(ctx context.Context, b ...*BatchElem) error
}

type Notifier interface {
	Notify(ctx context.Context, method string, params any) error
}

type Conn interface {
	Doer
	BatchCaller
	io.Closer
}

type StreamingConn interface {
	Conn
	Notifier
}
