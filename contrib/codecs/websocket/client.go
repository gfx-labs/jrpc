package websocket

import (
	"github.com/gfx-labs/jrpc/contrib/codecs/rdwr"
	"github.com/gfx-labs/jrpc/pkg/jsonrpc"

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
	if err := c.Client.Close(); err != nil {
		return err
	}
	if err := c.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
		return err
	}
	return nil
}
