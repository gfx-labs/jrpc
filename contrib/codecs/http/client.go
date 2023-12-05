package http

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"golang.org/x/net/http2"

	"gfx.cafe/open/jrpc/pkg/jjson"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

var (
	ErrClientQuit                = errors.New("client is closed")
	ErrNoResult                  = errors.New("no result in JSON-RPC response")
	ErrSubscriptionQueueOverflow = errors.New("subscription queue overflow")
	errClientReconnected         = errors.New("client reconnected")
	errDead                      = errors.New("connection lost")
)

var DefaultH2CClient = &http.Client{
	Transport: &http2.Transport{
		// So http2.Transport doesn't complain the URL scheme isn't 'https'
		AllowHTTP: true,
		// Pretend we are dialing a TLS endpoint. (Note, we ignore the passed tls.Config)
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		},
	},
}

var _ jsonrpc.Conn = (*Client)(nil)

// Client represents a connection to an RPC server.
type Client struct {
	remote string
	c      *http.Client

	id atomic.Int64

	headers http.Header

	m       jsonrpc.Middlewares
	handler jsonrpc.Handler
	mu      sync.RWMutex
}

func (c *Client) Mount(h jsonrpc.Middleware) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = append(c.m, h)
	c.handler = c.m.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		// do nothing on no handler
	})
}

func DialHTTP(target string) (*Client, error) {
	return Dial(nil, http.DefaultClient, target)
}
func DialH2C(target string) (*Client, error) {
	return Dial(nil, DefaultH2CClient, target)
}
func Dial(ctx context.Context, client *http.Client, target string) (*Client, error) {
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{remote: target, c: client, headers: http.Header{
		"Content-Type": []string{"application/json"},
	}}, nil
}

func (c *Client) SetHeader(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.headers.Set(key, value)
}

func (c *Client) Do(ctx context.Context, result any, method string, params any) error {
	req, err := jsonrpc.NewRequest(ctx, jsonrpc.NewId(c.id.Add(1)), method, params)
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
		return &jsonrpc.HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       b,
		}
	}
	msg := &jsonrpc.Message{}

	err = json.NewDecoder(resp.Body).Decode(&msg)
	if err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	if msg.Error != nil {
		return msg.Error
	}
	if result != nil && msg.Result != nil {
		err = json.NewDecoder(msg.Result).Decode(result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Notify(ctx context.Context, method string, params any) error {
	req, err := jsonrpc.NewRequest(ctx, nil, method, params)
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

func (c *Client) Close() error {
	return nil
}

func (c *Client) Closed() <-chan struct{} {
	return make(chan struct{})
}

func (c *Client) post(req *jsonrpc.Request) (*http.Response, error) {
	// TODO: use buffer for this
	buf := jjson.GetBuf()
	defer jjson.PutBuf(buf)
	buf.Reset()
	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.postBuf(req.Context(), buf)
	if err != nil {
		return nil, err
	}
	return resp, nil
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
	return c.c.Do(hreq)
}
