package rdwr

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"

	"gfx.cafe/open/jrpc/pkg/clientutil"
	"gfx.cafe/open/jrpc/pkg/codec"
)

type Client struct {
	p *clientutil.IdReply

	rd io.Reader
	wr io.Writer

	ctx context.Context
	cn  context.CancelFunc

	m       codec.Middlewares
	handler codec.Handler
	mu      sync.RWMutex

	handlerPeer codec.PeerInfo
}

func NewClient(rd io.Reader, wr io.Writer) *Client {
	cl := &Client{
		p:  clientutil.NewIdReply(),
		rd: rd,
		wr: wr,
		handlerPeer: codec.PeerInfo{
			Transport:  "ipc",
			RemoteAddr: "",
		},
		handler: codec.HandlerFunc(func(w codec.ResponseWriter, r *codec.Request) {}),
	}
	cl.ctx, cl.cn = context.WithCancel(context.Background())
	go cl.listen()
	return cl
}

func (c *Client) SetHandlerPeer(pi codec.PeerInfo) {
	c.handlerPeer = pi
}

func (c *Client) Mount(h codec.Middleware) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = append(c.m, h)
	c.handler = c.m.HandlerFunc(func(w codec.ResponseWriter, r *codec.Request) {
		// do nothing on no handler
	})
}

func (c *Client) listen() error {
	var msg json.RawMessage
	for {
		err := json.NewDecoder(c.rd).Decode(&msg)
		if err != nil {
			return err
		}
		msgs, _ := codec.ParseMessage(msg)
		for i := range msgs {
			v := msgs[i]
			if v == nil {
				continue
			}
			id := v.ID
			//  messages without ids are notifications
			if id == nil {
				var handler codec.Handler
				c.mu.RLock()
				handler = c.handler
				c.mu.RUnlock()
				// writer should only be allowed to send notifications
				// reader should contain the message above
				// the context is the client context
				req := codec.NewRawRequest(c.ctx,
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
	req, err := codec.NewRequest(ctx, codec.NewId(id), method, params)
	if err != nil {
		return err
	}
	fwd, err := json.Marshal(req)
	if err != nil {
		return err
	}
	err = c.writeContext(req.Context(), fwd)
	if err != nil {
		return err
	}
	ans, err := c.p.Ask(req.Context(), *id)
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
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	reqs := make([]*codec.Request, 0, len(b))
	ids := make([]*codec.ID, 0, len(b))
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
	err = c.writeContext(ctx, buf.Bytes())
	if err != nil {
		return err
	}
	// TODO: wait for response
	wg := sync.WaitGroup{}
	wg.Add(len(ids))
	for i := range ids {
		idx := i
		go func() {
			defer wg.Done()
			ans, err := c.p.Ask(reqs[idx].Context(), *ids[idx])
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
	return c.writeContext(ctx, fwd)
}

func (c *Client) SetHeader(key string, value string) {
}

func (c *Client) Close() error {
	c.cn()
	return nil
}

func (c *Client) writeContext(ctx context.Context, xs []byte) error {
	errch := make(chan error)
	go func() {
		_, err := c.wr.Write(xs)
		select {
		case errch <- err:
		case <-ctx.Done():
		case <-c.ctx.Done():
		}
	}()
	select {
	case err := <-errch:
		return err
	case <-c.ctx.Done():
		return c.ctx.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}
