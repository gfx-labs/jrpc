package client

import (
	"context"
	"sync"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/pkg/codec"
)

type Reconnecting struct {
	dialer     func(ctx context.Context) (jrpc.Conn, error)
	base       codec.Conn
	alive      bool
	middleware []codec.Middleware

	mu sync.Mutex
}

func NewReconnecting(dialer func(ctx context.Context) (jrpc.Conn, error)) *Reconnecting {
	r := &Reconnecting{
		dialer: dialer,
	}
	return r
}

func (r *Reconnecting) getClient(ctx context.Context) (jrpc.Conn, error) {
	reconnect := func() error {
		conn, err := r.dialer(ctx)
		if err != nil {
			return err
		}
		r.base = conn
		r.alive = true
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.base == nil {
		err := reconnect()
		if err != nil {
			return nil, err
		}
	} else {
		select {
		case <-r.base.Closed():
			err := reconnect()
			if err != nil {
				return nil, err
			}
		default:
		}
	}
	return r.base, nil
}

func (r *Reconnecting) Do(ctx context.Context, result any, method string, params any) error {
	errChan := make(chan error)
	go func() {
		conn, err := r.getClient(ctx)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- conn.Do(ctx, result, method, params)
	}()
	return <-errChan
}

func (r *Reconnecting) BatchCall(ctx context.Context, b ...*codec.BatchElem) error {
	errChan := make(chan error)
	go func() {
		conn, err := r.getClient(ctx)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- conn.BatchCall(ctx, b...)
	}()
	return <-errChan
}

func (r *Reconnecting) Mount(m codec.Middleware) {
	r.middleware = append(r.middleware, m)
}

// why would you want to do this....
func (r *Reconnecting) Close() error {
	return nil
}

// never....
func (r *Reconnecting) Closed() <-chan struct{} {
	return make(<-chan struct{})
}
