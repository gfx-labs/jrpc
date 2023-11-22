package websocket

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"gfx.cafe/open/websocket"

	_ "net/http/pprof"

	"gfx.cafe/open/jrpc/pkg/jjson"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/serverutil"
)

type Codec struct {
	closed chan struct{}
	conn   *websocket.Conn

	currentFrame io.WriteCloser
	wrLock       sync.Mutex

	decBuf  json.RawMessage
	decLock sync.Mutex

	i jsonrpc.PeerInfo
}

func newWebsocketCodec(ctx context.Context, conn *websocket.Conn, host string, req http.Header) *Codec {
	conn.SetReadLimit(WsMessageSizeLimit)
	c := &Codec{
		closed: make(chan struct{}),
		conn:   conn,
	}
	c.i.Transport = "ws"
	// Fill in connection details.
	c.i.HTTP.Host = host
	// traefik proxy protocol headers
	c.i.HTTP.Origin = req.Get("X-Real-Ip")
	if c.i.HTTP.Origin == "" {
		c.i.HTTP.Origin = req.Get("X-Forwarded-For")
	}
	// origin header fallback
	if c.i.HTTP.Origin == "" {
		c.i.HTTP.Origin = req.Get("origin")
	}
	c.i.RemoteAddr = c.i.HTTP.Origin
	c.i.HTTP.UserAgent = req.Get("User-Agent")
	c.i.HTTP.Headers = req
	// Start pinger.
	go heartbeat(ctx, conn, WsPingInterval)
	return c
}

func heartbeat(ctx context.Context, c *websocket.Conn, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		err := c.Ping(ctx)
		if err != nil {
			return
		}
		t.Reset(time.Minute)
	}
}

func (c *Codec) decodeSingleMessage(ctx context.Context) (*serverutil.Bundle, error) {
	c.decLock.Lock()
	defer c.decLock.Unlock()
	c.decBuf = c.decBuf[:0]
	_, r, err := c.conn.Reader(ctx)
	if err != nil {
		return nil, err
	}
	defer io.Copy(io.Discard, r)
	err = jjson.Decode(r, &c.decBuf)
	if err != nil {
		return nil, err
	}
	return serverutil.ParseBundle(c.decBuf), nil
}

func (c *Codec) ReadBatch(ctx context.Context) ([]*jsonrpc.Message, bool, error) {
	ans, err := c.decodeSingleMessage(ctx)
	if err != nil {
		return nil, false, err
	}
	return ans.Messages, ans.Batch, nil
}

func (c *Codec) Write(p []byte) (n int, err error) {
	c.wrLock.Lock()
	defer c.wrLock.Unlock()
	if c.currentFrame == nil {
		wr, err := c.conn.Writer(context.Background(), websocket.MessageText)
		if err != nil {
			return 0, err
		}
		c.currentFrame = wr
	}
	return c.currentFrame.Write(p)
}

func (c *Codec) Flush() error {
	c.wrLock.Lock()
	defer c.wrLock.Unlock()
	if c.currentFrame == nil {
		wr, err := c.conn.Writer(context.Background(), websocket.MessageText)
		if err != nil {
			return err
		}
		return wr.Close()
	}
	err := c.currentFrame.Close()
	if err != nil {
		return err
	}
	c.currentFrame = nil
	return nil
}

func (c *Codec) PeerInfo() jsonrpc.PeerInfo {
	return c.i
}

func (c *Codec) Closed() <-chan struct{} {
	return c.closed
}

func (c *Codec) Close() error {
	select {
	case <-c.closed:
		return nil
	default:
		close(c.closed)
	}
	return c.conn.Close(websocket.StatusNormalClosure, "")
}

func (c *Codec) RemoteAddr() string {
	return c.i.RemoteAddr
}

var _ jsonrpc.ReaderWriter = (*Codec)(nil)
