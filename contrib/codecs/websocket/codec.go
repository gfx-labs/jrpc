package websocket

import (
	"context"
	"net/http"
	"time"

	"nhooyr.io/websocket"

	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"
	"gfx.cafe/open/jrpc/pkg/codec"
)

type Codec struct {
	*rdwr.Codec
	conn *websocket.Conn

	i codec.PeerInfo
}

func newWebsocketCodec(ctx context.Context, conn *websocket.Conn, host string, req http.Header) *Codec {
	conn.SetReadLimit(WsMessageSizeLimit)
	netConn := websocket.NetConn(ctx, conn, websocket.MessageText)
	c := &Codec{
		Codec: rdwr.NewCodec(netConn, netConn, nil),
		conn:  conn,
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

func (c *Codec) PeerInfo() codec.PeerInfo {
	return c.i
}

func (c *Codec) Close() error {
	if err := c.Codec.Close(); err != nil {
		return err
	}
	return c.conn.Close(websocket.StatusNormalClosure, "")
}

func (c *Codec) RemoteAddr() string {
	return c.i.RemoteAddr
}
