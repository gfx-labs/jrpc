package websocket

import (
	"context"
	jrpc2 "gfx.cafe/open/jrpc/pkg/codec"
	"sync"

	"gfx.cafe/open/jrpc"
	"nhooyr.io/websocket"
)

type Client struct {
	conn          *websocket.Conn
	reconnectFunc reconnectFunc

	mu sync.RWMutex
}

type reconnectFunc func(ctx context.Context) (*websocket.Conn, error)

func newClient(initctx context.Context, connect reconnectFunc) (*Client, error) {
	conn, err := connect(initctx)
	if err != nil {
		return nil, err
	}
	c := &Client{}
	c.conn = conn
	c.reconnectFunc = connect
	return c, nil
}

func (c *Client) Do(ctx context.Context, result any, method string, params any) error {
	panic("not implemented") // TODO: Implement
}

func (c *Client) BatchCall(ctx context.Context, b ...jrpc2.BatchElem) error {
	panic("not implemented") // TODO: Implement
}

func (c *Client) SetHeader(key string, value string) {
	panic("not implemented") // TODO: Implement
}

func (c *Client) Close() error {
	panic("not implemented") // TODO: Implement
}

func (c *Client) Notify(ctx context.Context, method string, args ...any) error {
	panic("not implemented") // TODO: Implement
}

func (c *Client) Subscribe(ctx context.Context, namespace string, channel any, args ...any) (*jrpc.ClientSubscription, error) {
	panic("not implemented") // TODO: Implement
}
