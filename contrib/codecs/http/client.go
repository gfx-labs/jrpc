package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"gfx.cafe/open/jrpc/pkg/codec"

	"gfx.cafe/open/jrpc/pkg/clientutil"
	"gfx.cafe/util/go/bufpool"
)

var (
	ErrClientQuit                = errors.New("client is closed")
	ErrNoResult                  = errors.New("no result in JSON-RPC response")
	ErrSubscriptionQueueOverflow = errors.New("subscription queue overflow")
	errClientReconnected         = errors.New("client reconnected")
	errDead                      = errors.New("connection lost")
)

const (
	// Timeouts
	defaultDialTimeout = 10 * time.Second // used if context has no deadline
	subscribeTimeout   = 5 * time.Second  // overall timeout eth_subscribe, rpc_modules calls
)

var _ codec.Conn = (*Client)(nil)

// Client represents a connection to an RPC server.
type Client struct {
	remote string
	c      *http.Client

	id atomic.Int64

	headers http.Header
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
	c.headers.Set(key, value)
}

func (c *Client) Do(ctx context.Context, result any, method string, params any) error {
	req := codec.NewRequestInt(ctx, int(c.id.Add(1)), method, params)
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
	// TODO: this can be reused
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

func (c *Client) post(req *codec.Request) (*http.Response, error) {
	//TODO: use buffer for this
	buf := bufpool.GetStd()
	defer bufpool.PutStd(buf)
	buf.Reset()
	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		return nil, err
	}
	hreq, err := http.NewRequestWithContext(req.Context(), http.MethodPost, c.remote, buf)
	if err != nil {
		return nil, err
	}
	for k, v := range c.headers {
		for _, vv := range v {
			hreq.Header.Add(k, vv)
		}
	}
	return c.c.Do(hreq)
}

func (c *Client) Notify(ctx context.Context, method string, params any) error {
	req := codec.NewNotification(ctx, method, params)
	resp, err := c.post(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return err
}

func (c *Client) BatchCall(ctx context.Context, b ...*codec.BatchElem) error {
	reqs := make([]*codec.Request, len(b))
	ids := make([]int, 0, len(b))
	for _, v := range b {
		if v.IsNotification {
			reqs = append(reqs, codec.NewRequest(ctx, "", v.Method, v.Params))
		} else {
			id := int(c.id.Add(1))
			ids = append(ids, id)
			reqs = append(reqs, codec.NewRequestInt(ctx, id, v.Method, v.Params))
		}
	}
	dat, err := json.Marshal(reqs)
	if err != nil {
		return err
	}
	resp, err := c.c.Post(c.remote, "application/json", bytes.NewBuffer(dat))
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
