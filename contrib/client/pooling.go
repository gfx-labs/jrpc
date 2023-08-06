package client

import (
	"context"
	"net"
	"sync"
	"sync/atomic"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/pkg/codec"
)

var _ codec.Conn = (*Pooling)(nil)

type Pooling struct {
	dialer     func(ctx context.Context) (jrpc.Conn, error)
	conns      chan codec.Conn
	base       codec.Conn
	closed     atomic.Bool
	middleware []codec.Middleware

	mu sync.Mutex
}

func NewPooling(dialer func(ctx context.Context) (jrpc.Conn, error), max int) *Pooling {
	r := &Pooling{
		dialer: dialer,
		conns:  make(chan codec.Conn, max),
	}
	return r
}

func (r *Pooling) Do(ctx context.Context, result any, method string, params any) error {
	if r.closed.Load() {
		return net.ErrClosed
	}
	errChan := make(chan error)
	go func() {
		conn, err := r.getClient(ctx)
		if err != nil {
			errChan <- err
			return
		}
		defer r.putClient(conn)
		errChan <- conn.Do(ctx, result, method, params)
	}()
	return <-errChan
}

func (r *Pooling) BatchCall(ctx context.Context, b ...*codec.BatchElem) error {
	if r.closed.Load() {
		return net.ErrClosed
	}
	errChan := make(chan error)
	go func() {
		conn, err := r.getClient(ctx)
		if err != nil {
			errChan <- err
			return
		}
		defer r.putClient(conn)
		errChan <- conn.BatchCall(ctx, b...)
	}()
	return <-errChan
}

func (p *Pooling) Mount(m codec.Middleware) {
	p.middleware = append(p.middleware, m)
}

func (p *Pooling) Close() error {
	if p.closed.CompareAndSwap(false, true) {
		for k := range p.conns {
			k.Close()
		}
	}
	return nil
}

func (p *Pooling) Closed() <-chan struct{} {
	return make(<-chan struct{})
}

func (r *Pooling) getClient(ctx context.Context) (jrpc.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-r.conns:
	default:
	}
	conn, err := r.dialer(ctx)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
func (r *Pooling) putClient(conn jrpc.Conn) {
	if r.closed.Load() {
		return
	}
	select {
	case <-conn.Closed():
	default:
	}
	select {
	case r.conns <- conn:
	default:
		conn.Close()
	}
}
