package inproc

import (
	"context"
	"encoding/json"
	"sync"

	"gfx.cafe/open/jrpc/pkg/clientutil"
	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/serverutil"
)

type Client struct {
	p *clientutil.IdReply
	c *Codec

	ctx context.Context
	cn  context.CancelFunc

	m       codec.Middlewares
	handler codec.Handler
	mu      sync.Mutex
}

func (c *Client) Mount(h codec.Middleware) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = append(c.m, h)
	c.handler = c.m.HandlerFunc(func(w codec.ResponseWriter, r *codec.Request) {
		// do nothing on no handler
	})
}

func NewClient(c *Codec, handler codec.Handler) *Client {
	cl := &Client{
		p:       clientutil.NewIdReply(),
		c:       c,
		handler: handler,
	}
	cl.ctx, cl.cn = context.WithCancel(context.Background())
	go cl.listen()

	return cl
}

func (c *Client) Closed() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Client) listen() error {
	var msg json.RawMessage
	defer c.cn()
	for {
		err := json.NewDecoder(c.c.rd).Decode(&msg)
		if err != nil {
			return err
		}
		msgs, _ := codec.ParseMessage(msg)
		for i := range msgs {
			v := msgs[i]
			id := v.ID
			if id == nil {
				if c.handler != nil {
					req := codec.NewRawRequest(c.c.ctx,
						nil,
						v.Method,
						v.Params,
					)
					req.Peer = codec.PeerInfo{
						Transport:  "ipc",
						RemoteAddr: "",
					}
					c.handler.ServeRPC(nil, req)
				}
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
	if ctx == nil {
		ctx = context.Background()
	}
	dat, err := json.Marshal(params)
	if err != nil {
		return err
	}
	id := c.p.NextId()
	fwd := &serverutil.Bundle{
		Messages: []*codec.Message{{
			ID:     id,
			Method: method,
			Params: dat,
		}},
		Batch: false,
	}
	select {
	case c.c.msgs <- fwd:
	case <-ctx.Done():
		return ctx.Err()
	}
	ans, err := c.p.Ask(ctx, *id)
	if err != nil {
		return err
	}
	if result != nil {
		err = json.Unmarshal(ans, result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) BatchCall(ctx context.Context, b ...*codec.BatchElem) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ids := make([]*codec.ID, 0, len(b))
	reqs := &serverutil.Bundle{Batch: true}
	for _, v := range b {
		id := c.p.NextId()
		dat, err := json.Marshal(v.Params)
		if err != nil {
			return err
		}
		req := &codec.Message{ID: id, Method: v.Method, Params: dat}
		ids = append(ids, id)
		reqs.Messages = append(reqs.Messages, req)
	}
	c.c.msgs <- reqs
	// TODO: wait for response
	wg := sync.WaitGroup{}
	wg.Add(len(ids))
	for i := range ids {
		idx := i
		go func() {
			defer wg.Done()
			ans, err := c.p.Ask(ctx, *ids[idx])
			if err != nil {
				b[idx].Error = err
				return
			}
			if b[idx].Result != nil {
				err = json.Unmarshal(ans, b[idx].Result)
				if err != nil {
					b[idx].Error = err
					return
				}
			}
		}()
	}
	wg.Wait()

	return nil
}

func (c *Client) SetHeader(key string, value string) {
}

func (c *Client) Close() error {
	return c.c.Close()
}

func (c *Client) Notify(ctx context.Context, method string, params any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	dat, err := json.Marshal(params)
	if err != nil {
		return err
	}
	msg := &serverutil.Bundle{
		Messages: []*codec.Message{{
			ID:     nil,
			Method: method,
			Params: dat,
		}},
		Batch: false,
	}
	select {
	case c.c.msgs <- msg:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
