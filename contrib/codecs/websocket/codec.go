package websocket

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"gfx.cafe/open/websocket"
	"golang.org/x/sync/semaphore"

	"gfx.cafe/open/jrpc/pkg/jjson"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

type Codec struct {
	closed chan struct{}
	conn   *websocket.Conn
	closer func()
	ctx    context.Context

	currentFrame io.WriteCloser
	wrLock       sync.Mutex

	decBuf  json.RawMessage
	decLock *semaphore.Weighted

	i jsonrpc.PeerInfo
}

func newWebsocketCodec(ctx context.Context, conn *websocket.Conn, host string, req *http.Request) *Codec {
	conn.SetReadLimit(WsMessageSizeLimit)

	ctx, cn := context.WithCancel(ctx)
	c := &Codec{
		closed:  make(chan struct{}),
		conn:    conn,
		decLock: semaphore.NewWeighted(1),
		ctx:     ctx,
	}
	c.closer = func() {
		cn()
	}
	c.i.Transport = "ws"
	// Fill in connection details.
	c.i.HTTP = req.Clone(ctx)
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

func (c *Codec) ReadBatch(ctx context.Context) ([]*jsonrpc.Message, bool, error) {
	if err := c.decLock.Acquire(ctx, 1); err != nil {
		return nil, false, err
	}
	defer c.decLock.Release(1)
	c.decBuf = c.decBuf[:0]
	_, r, err := c.conn.Reader(ctx)
	if err != nil {
		return nil, false, err
	}
	defer io.Copy(io.Discard, r)
	err = jjson.Decode(r, &c.decBuf)
	if err != nil {
		return nil, false, err
	}
	a, b := jsonrpc.ParseMessage(c.decBuf)
	return a, b, nil
}

func (c *Codec) Write(p []byte) (n int, err error) {
	c.wrLock.Lock()
	defer c.wrLock.Unlock()
	if c.currentFrame == nil {
		wr, err := c.conn.Writer(c.ctx, websocket.MessageText)
		if err != nil {
			c.Close()
			return 0, err
		}
		c.currentFrame = wr
	}

	n, err = c.currentFrame.Write(p)
	if err != nil {
		c.Close()
	}
	return
}

func (c *Codec) Flush() error {
	c.wrLock.Lock()
	defer c.wrLock.Unlock()
	if c.currentFrame == nil {
		return nil
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
		c.closer()
		close(c.closed)
	}
	return c.conn.Close(websocket.StatusNormalClosure, "")
}

func (c *Codec) RemoteAddr() string {
	return c.i.RemoteAddr
}

var _ jsonrpc.ReaderWriter = (*Codec)(nil)
