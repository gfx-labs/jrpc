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

	handler codec.Handler
}

func NewClient(rd io.Reader, wr io.Writer, handler codec.Handler) *Client {
	cl := &Client{
		p:       clientutil.NewIdReply(),
		rd:      rd,
		wr:      wr,
		handler: handler,
	}
	cl.ctx, cl.cn = context.WithCancel(context.Background())
	go cl.listen()
	return cl
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
			id := v.ID.Number()
			//  messages without ids are notifications
			if id == 0 {
				//TODO: implement this
				if c.handler != nil {
					// writer should only be allowed to send notifications
					// reader should contain the message above
					// the context is the client context
					c.handler.ServeRPC(nil, codec.NewRequestFromRaw(c.ctx, &codec.RequestMarshaling{
						Method: v.Method,
						Params: v.Params,
						Peer: codec.PeerInfo{
							Transport:  "ipc",
							RemoteAddr: "",
						},
					}))
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
	if ctx == nil {
		ctx = context.Background()
	}
	id := c.p.NextId()
	req := codec.NewRequestInt(ctx, id, method, params)
	fwd, err := json.Marshal(req)
	if err != nil {
		return err
	}
	err = c.writeContext(ctx, fwd)
	if err != nil {
		return err
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
	if ctx == nil {
		ctx = context.Background()
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	reqs := make([]*codec.Request, 0, len(b))
	ids := make([]int, 0, len(b))
	for _, v := range b {
		id := c.p.NextId()
		req := codec.NewRequestInt(ctx, id, v.Method, v.Params)
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
func (c *Client) Notify(ctx context.Context, method string, params any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	req := codec.NewRequest(ctx, "", method, params)
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
