package inproc

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"

	"gfx.cafe/open/jrpc/pkg/clientutil"
	"gfx.cafe/open/jrpc/pkg/codec"
)

type Client struct {
	p *clientutil.IdReply
	c *Codec

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
	go cl.listen()
	return cl
}

func (c *Client) listen() error {
	var msg json.RawMessage
	for {
		err := json.NewDecoder(c.c.rd).Decode(&msg)
		if err != nil {
			return err
		}
		msgs, _ := codec.ParseMessage(msg)
		for i := range msgs {
			v := msgs[i]
			id := v.ID.Number()
			if id == 0 {
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
			c.p.Resolve(id, v.Result, err)
		}
	}

}

func (c *Client) Do(ctx context.Context, result any, method string, params any) error {
	id := c.p.NextId()
	req, err := codec.NewRequest(ctx, codec.NewId(id), method, params)
	if err != nil {
		return err
	}
	fwd, err := json.Marshal(req)
	if err != nil {
		return err
	}
	select {
	case c.c.msgs <- fwd:
	case <-req.Context().Done():
		return req.Context().Err()
	}
	ans, err := c.p.Ask(req.Context(), id)
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
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	reqs := make([]*codec.Request, 0, len(b))
	ids := make([]int, 0, len(b))
	for _, v := range b {
		id := c.p.NextId()
		req, err := codec.NewRequest(ctx, codec.NewId(id), v.Method, v.Params)
		if err != nil {
			return err
		}
		ids = append(ids, id)
		reqs = append(reqs, req)
	}
	err := enc.Encode(reqs)
	if err != nil {
		return err
	}
	c.c.msgs <- buf.Bytes()
	// TODO: wait for response
	wg := sync.WaitGroup{}
	wg.Add(len(ids))
	for i := range ids {
		idx := i
		go func() {
			defer wg.Done()
			ans, err := c.p.Ask(reqs[idx].Context(), ids[idx])
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

	return err
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
	req, err := codec.NewRequest(ctx, nil, method, params)
	if err != nil {
		return err
	}
	fwd, err := json.Marshal(req)
	if err != nil {
		return err
	}
	select {
	case c.c.msgs <- fwd:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
