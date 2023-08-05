package codecs

import (
	"context"
	"net"
	"net/url"

	"gfx.cafe/open/jrpc/contrib/codecs/http"
	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"
	"gfx.cafe/open/jrpc/contrib/codecs/websocket"
	"gfx.cafe/open/jrpc/exp/redis"
	"gfx.cafe/open/jrpc/pkg/codec"
)

func DialContext(ctx context.Context, u string) (codec.Conn, error) {
	pu, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	switch pu.Scheme {
	case "http", "https":
		return http.Dial(ctx, nil, u)
	case "ws", "wss":
		return websocket.DialWebsocket(ctx, u, "")
	case "redis":
		domain := pu.Query().Get("domain")
		if domain == "" {
			domain = "jrpc"
		}
		return redis.Dial(pu.Host, domain), nil
	case "tcp":
		tcpAddr, err := net.ResolveTCPAddr("tcp", u)
		if err != nil {
			return nil, err
		}
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			return nil, err
		}
		return rdwr.NewClient(conn, conn), nil
	}
	return nil, nil
}

func Dial(u string) (codec.Conn, error) {
	ctx := context.Background()
	return DialContext(ctx, u)
}
