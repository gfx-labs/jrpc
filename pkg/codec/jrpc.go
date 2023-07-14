package codec

import (
	"context"
	"io"
)

// http.handler, but for jrpc
type Handler interface {
	ServeRPC(w ResponseWriter, r *Request)
}

// http.HandlerFunc,but for jrpc
type HandlerFunc func(w ResponseWriter, r *Request)

func (fn HandlerFunc) ServeRPC(w ResponseWriter, r *Request) {
	(fn)(w, r)
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

type Conn interface {
	Doer
	BatchCaller
	io.Closer

	Mounter
}

type StreamingConn interface {
	Conn
	Notifier
}
