package websocket

import (
	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"

	"context"

	"gfx.cafe/open/websocket"
)

type Client struct {
	*rdwr.Client
	conn *websocket.Conn
}

func newClient(conn *websocket.Conn) (*Client, error) {
	conn.SetReadLimit(WsMessageSizeLimit)
	netConn := websocket.NetConn(context.Background(), conn, websocket.MessageText)
	c := &Client{
		Client: rdwr.NewClient(netConn, netConn),
		conn:   conn,
	}
	c.SetHandlerPeer(jsonrpc.PeerInfo{
		Transport:  "ws",
		RemoteAddr: "",
	})
	return c, nil
}

func (c *Client) Close() error {
	return c.conn.Close(websocket.StatusNormalClosure, "")
}
