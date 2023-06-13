package websocket

import (
	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"

	"context"

	"nhooyr.io/websocket"
)

type Client struct {
	*rdwr.Client
	conn *websocket.Conn
}

func newClient(conn *websocket.Conn) (*Client, error) {
	conn.SetReadLimit(WsMessageSizeLimit)
	netConn := websocket.NetConn(context.Background(), conn, websocket.MessageText)
	c := &Client{
		Client: rdwr.NewClient(netConn, netConn, nil),
		conn:   conn,
	}
	return c, nil
}

func (c *Client) Close() error {
	return c.conn.Close(websocket.StatusNormalClosure, "")
}
