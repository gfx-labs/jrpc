package jrpc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/goccy/go-json"
)

func (hc *httpConn) doRequest(ctx context.Context, msg any) (io.ReadCloser, error) {
	// TODO:
	// the jsoniter encoder performs a lot better here, not sure why. (nearly 10%? maybe more)
	body, err := jzon.Marshal(msg)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", hc.url, io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return nil, err
	}
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }

	// set headers
	hc.mu.Lock()
	req.Header = hc.headers.Clone()
	hc.mu.Unlock()

	// do request
	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var buf bytes.Buffer
		var body []byte
		if _, err := buf.ReadFrom(resp.Body); err == nil {
			body = buf.Bytes()
		}
		return nil, HTTPError{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}
	return resp.Body, nil
}

// DialHTTPWithClient creates a new RPC client that connects to an RPC server over HTTP
// using the provided HTTP Client.
func DialHTTPWithClient(endpoint string, client *http.Client) (*Client, error) {
	// Sanity check URL so we don't end up with a client that will fail every request.
	_, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	initctx := context.Background()
	headers := make(http.Header, 2)
	headers.Set("accept", contentType)
	headers.Set("content-type", contentType)
	return newClient(initctx, func(context.Context) (ServerCodec, error) {
		hc := &httpConn{
			client:  client,
			headers: headers,
			url:     endpoint,
			closeCh: make(chan any),
		}
		return hc, nil
	})
}

// DialHTTP creates a new RPC client that connects to an RPC server over HTTP.
func DialHTTP(endpoint string) (*Client, error) {
	return DialHTTPWithClient(endpoint, new(http.Client))
}

func (c *Client) sendHTTP(ctx context.Context, op *requestOp, msg any) error {
	hc := c.writeConn.(*httpConn)
	respBody, err := hc.doRequest(ctx, msg)
	if err != nil {
		return err
	}
	defer respBody.Close()

	var respmsg jsonrpcMessage
	if err := json.NewDecoder(respBody).Decode(&respmsg); err != nil {
		return err
	}
	op.resp <- &respmsg
	return nil
}

func (c *Client) sendBatchHTTP(ctx context.Context, op *requestOp, msgs []*jsonrpcMessage) error {
	hc := c.writeConn.(*httpConn)
	respBody, err := hc.doRequest(ctx, msgs)
	if err != nil {
		return err
	}
	defer respBody.Close()
	var respmsgs []jsonrpcMessage
	if err := json.NewDecoder(respBody).Decode(&respmsgs); err != nil {
		return err
	}
	for i := 0; i < len(respmsgs); i++ {
		op.resp <- &respmsgs[i]
	}
	return nil
}

type httpConn struct {
	client    *http.Client
	url       string
	closeOnce sync.Once
	closeCh   chan any
	mu        sync.Mutex // protects headers
	headers   http.Header
}

// httpConn implements ServerCodec, but it is treated specially by Client
// and some methods don't work. The panic() stubs here exist to ensure
// this special treatment is correct.
func (hc *httpConn) WriteJSON(context.Context, any) error {
	panic("WriteJSON called on httpConn")
}

func (hc *httpConn) PeerInfo() PeerInfo {
	panic("PeerInfo called on httpConn")
}

func (hc *httpConn) RemoteAddr() string {
	return hc.url
}

func (hc *httpConn) ReadBatch() ([]*jsonrpcMessage, bool, error) {
	<-hc.closeCh
	return nil, false, io.EOF
}

func (hc *httpConn) Close() error {
	hc.closeOnce.Do(func() { close(hc.closeCh) })
	return nil
}

func (hc *httpConn) Closed() <-chan any {
	return hc.closeCh
}
