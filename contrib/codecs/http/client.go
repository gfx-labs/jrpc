package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	"gfx.cafe/open/jrpc/pkg/codec"

	"gfx.cafe/util/go/bufpool"

	"gfx.cafe/open/jrpc/pkg/clientutil"
)

var (
	ErrClientQuit                = errors.New("client is closed")
	ErrNoResult                  = errors.New("no result in JSON-RPC response")
	ErrSubscriptionQueueOverflow = errors.New("subscription queue overflow")
	errClientReconnected         = errors.New("client reconnected")
	errDead                      = errors.New("connection lost")
)

var _ codec.Conn = (*Client)(nil)

// Client represents a connection to an RPC server.
type Client struct {
	remote string
	c      *http.Client

	id atomic.Int64

	headers http.Header

	m       codec.Middlewares
	handler codec.Handler
	mu      sync.RWMutex
}

func (c *Client) Mount(h codec.Middleware) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = append(c.m, h)
	c.handler = c.m.HandlerFunc(func(w codec.ResponseWriter, r *codec.Request) {
		// do nothing on no handler
	})
}

func DialHTTP(target string) (*Client, error) {
	return Dial(nil, http.DefaultClient, target)
}

func Dial(ctx context.Context, client *http.Client, target string) (*Client, error) {
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{remote: target, c: client, headers: http.Header{}}, nil
}

func (c *Client) SetHeader(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.headers.Set(key, value)
}

func (c *Client) Do(ctx context.Context, result any, method string, params any) error {
	req, err := codec.NewRequest(ctx, codec.NewId(c.id.Add(1)), method, params)
	if err != nil {
		return err
	}
	resp, err := c.post(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return &codec.HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       b,
		}
	}
	msg := clientutil.GetMessage()
	defer clientutil.PutMessage(msg)
	err = json.NewDecoder(resp.Body).Decode(&msg)
	if err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	if msg.Error != nil {
		return msg.Error
	}
	if result != nil && len(msg.Result) > 0 {
		err = json.Unmarshal(msg.Result, result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Notify(ctx context.Context, method string, params any) error {
	req, err := codec.NewRequest(ctx, nil, method, params)
	if err != nil {
		return err
	}
	resp, err := c.post(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return err
}

func (c *Client) BatchCall(ctx context.Context, b ...*codec.BatchElem) error {
	reqs := make([]*codec.Request, len(b))
	ids := make(map[int]int, len(b))
	for idx, v := range b {
		var rid *codec.ID
		if v.IsNotification {
		} else {
			id := int(c.id.Add(1))
			ids[idx] = id
			rid = codec.NewNumberIDPtr(int64(id))
		}
		req, err := codec.NewRequest(ctx, rid, v.Method, v.Params)
		if err != nil {
			return err
		}
		reqs = append(reqs, req)
	}
	dat, err := json.Marshal(reqs)
	if err != nil {
		return err
	}
	resp, err := c.postBuf(ctx, bytes.NewBuffer(dat))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	msgs := []*codec.Message{}
	for i := 0; i < len(ids); i++ {
		msg := clientutil.GetMessage()
		defer clientutil.PutMessage(msg)
		msgs = append(msgs, msg)
	}
	err = json.NewDecoder(resp.Body).Decode(&msgs)
	if err != nil {
		return err
	}
	clientutil.FillBatch(ids, msgs, b)
	return nil
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) Closed() <-chan struct{} {
	return make(chan struct{})
}

func (c *Client) post(req *codec.Request) (*http.Response, error) {
	// TODO: use buffer for this
	buf := bufpool.GetStd()
	defer bufpool.PutStd(buf)
	buf.Reset()
	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		return nil, err
	}
	return c.postBuf(req.Context(), buf)
}

func (c *Client) postBuf(ctx context.Context, rd io.Reader) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.remote, rd)
	if err != nil {
		return nil, err
	}
	func() {
		c.mu.RLock()
		defer c.mu.RUnlock()
		for k, v := range c.headers {
			for _, vv := range v {
				hreq.Header.Add(k, vv)
			}
		}
	}()
	hreq.Header.Add("Content-Type", "application/json")
	return c.c.Do(hreq)
}
