package rdwr

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"sync"

	"gfx.cafe/open/jrpc/pkg/clientutil"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/util/go/bufpool"
)

type Client struct {
	p *clientutil.IdReply

	rd io.Reader
	wr io.Writer

	ctx context.Context
	cn  context.CancelFunc

	m       jsonrpc.Middlewares
	handler jsonrpc.Handler
	writeCh chan struct{}

	mu sync.RWMutex

	handlerPeer jsonrpc.PeerInfo
}

func NewClient(rd io.Reader, wr io.Writer) *Client {
	cl := &Client{
		p:  clientutil.NewIdReply(),
		rd: bufio.NewReader(rd),
		wr: wr,
		handlerPeer: jsonrpc.PeerInfo{
			Transport:  "ipc",
			RemoteAddr: "",
		},
		handler: jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {}),
		writeCh: make(chan struct{}, 1),
	}
	cl.ctx, cl.cn = context.WithCancel(context.Background())
	go cl.listen()
	return cl
}

func (c *Client) SetHandlerPeer(pi jsonrpc.PeerInfo) {
	c.handlerPeer = pi
}

func (c *Client) Closed() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Client) Mount(h jsonrpc.Middleware) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = append(c.m, h)
	c.handler = c.m.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		// do nothing on no handler
	})
}

func (c *Client) listen() error {
	var msg json.RawMessage
	defer c.cn()
	dec := json.NewDecoder(bufio.NewReader(c.rd))
	for {
		err := dec.Decode(&msg)
		if err != nil {
			return err
		}
		msgs, _ := jsonrpc.ParseMessage(msg)
		for i := range msgs {
			v := msgs[i]
			if v == nil {
				continue
			}
			id := v.ID
			//  messages without ids are notifications
			if id == nil {
				var handler jsonrpc.Handler
				c.mu.RLock()
				handler = c.handler
				c.mu.RUnlock()
				// writer should only be allowed to send notifications
				// reader should contain the message above
				// the context is the client context
				req := jsonrpc.NewRawRequest(c.ctx,
					nil,
					v.Method,
					v.Params,
				)
				req.Peer = c.handlerPeer
				handler.ServeRPC(nil, req)
				continue
			}
			var err error
			if v.Error != nil {
				err = v.Error
			}
			c.p.Resolve(*id, v.Result, err)
		}
	}

}

func (c *Client) Do(ctx context.Context, result any, method string, params any) error {
	id := c.p.NextId()
	buf := bufpool.GetStd()
	defer bufpool.PutStd(buf)
	req, err := jsonrpc.NewRequest(ctx, jsonrpc.NewId(id), method, params)
	if err != nil {
		return err
	}
	err = json.NewEncoder(buf).Encode(req)
	if err != nil {
		return err
	}
	err = c.writeContext(req.Context(), buf.Bytes())
	if err != nil {
		return err
	}
	ans, err := c.p.Ask(req.Context(), *id)
	if err != nil {
		return err
	}
	if result != nil {
		err = json.NewDecoder(ans).Decode(result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Notify(ctx context.Context, method string, params any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := jsonrpc.NewRequest(ctx, nil, method, params)
	if err != nil {
		return err
	}
	fwd, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.writeContext(ctx, fwd)
}

func (c *Client) SetHeader(key string, value string) {
}

func (c *Client) Close() error {
	c.cn()
	return nil
}

func (c *Client) writeContext(ctx context.Context, xs []byte) error {
	select {
	case c.writeCh <- struct{}{}:
		defer func() {
			<-c.writeCh
		}()
		_, err := c.wr.Write(xs)
		return err
	case <-c.ctx.Done():
		return c.ctx.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}
