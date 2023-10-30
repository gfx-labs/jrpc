package client

import (
	"context"
	"net"
	"sync"
	"sync/atomic"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/contrib/extension/subscription"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

var _ jsonrpc.Conn = (*Pooling)(nil)
var _ subscription.Conn = (*Pooling)(nil)

type Pooling struct {
	dialer     func(ctx context.Context) (jrpc.Conn, error)
	conns      chan jsonrpc.Conn
	base       subscription.Conn
	closed     atomic.Bool
	middleware []jsonrpc.Middleware

	mu sync.Mutex
}

func (p *Pooling) getBase(ctx context.Context) (subscription.Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.base == nil {
		conn, err := subscription.UpgradeConn(p.dialer(ctx))
		if err != nil {
			return nil, err
		}
		p.base = conn
	}
	return p.base, nil
}

func (p *Pooling) Notify(ctx context.Context, method string, params any) error {
	base, err := p.getBase(ctx)
	if err != nil {
		return err
	}
	return base.Notify(ctx, method, params)
}

func (p *Pooling) Subscribe(ctx context.Context, namespace string, channel any, args any) (subscription.ClientSubscription, error) {
	base, err := p.getBase(ctx)
	if err != nil {
		return nil, err
	}
	return base.Subscribe(ctx, namespace, channel, args)
}

func NewPooling(ctx context.Context, dialer func(ctx context.Context) (jrpc.Conn, error), max int) (*Pooling, error) {
	r := &Pooling{
		dialer: dialer,
		conns:  make(chan jsonrpc.Conn, max),
	}

	return r, nil
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

func (r *Pooling) BatchCall(ctx context.Context, b ...*jsonrpc.BatchElem) error {
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

func (p *Pooling) Mount(m jsonrpc.Middleware) {
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
