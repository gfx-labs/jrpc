package inproc

import (
	"bytes"
	"context"
	"encoding/json"
	"gfx.cafe/open/jrpc/pkg/clientutil"
	"gfx.cafe/open/jrpc/pkg/codec"
	"sync"
)

type Client struct {
	p *clientutil.IdReply
	c *Codec

	handler codec.Handler
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
		for _, v := range msgs {
			id := v.ID.Number()
			if id == 0 {
				//if c.handler != nil {
				//	c.handler.ServeRPC(w, r)
				//}
				continue
			}
			c.p.Resolve(id, v.Result, v.Error)
		}
	}

}

func (c *Client) Do(ctx context.Context, result any, method string, params any) error {
	id := c.p.NextId()
	req := codec.NewRequestInt(ctx, id, method, params)
	fwd, err := json.Marshal(req)
	if err != nil {
		return err
	}
	c.c.msgs <- fwd
	ans, err := c.p.Ask(ctx, id)
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
		req := codec.NewRequestInt(ctx, id, v.Method, v.Params)
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
			ans, err := c.p.Ask(ctx, ids[idx])
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
	req := codec.NewRequest(ctx, "", method, params)
	fwd, err := json.Marshal(req)
	if err != nil {
		return err
	}
	c.c.msgs <- fwd
	return nil
}
