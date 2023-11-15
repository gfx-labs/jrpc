package codecs

import (
	"context"
	"net"
	"net/url"
	"strings"

	gohttp "net/http"

	"gfx.cafe/open/jrpc/contrib/codecs/http"
	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"
	"gfx.cafe/open/jrpc/contrib/codecs/websocket"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/server"
)

func init() {
	RegisterHandler(func(bind *url.URL, srv *server.Server, opts map[string]any) error {
		return gohttp.ListenAndServe(bind.Host, HttpHandler(srv))
	}, "http")
	RegisterHandler(func(bind *url.URL, srv *server.Server, opts map[string]any) error {
		origins := []string{}
		if val, ok := opts["origins"]; ok {
			if t, ok := val.([]string); ok {
				origins = t
			}
		}
		return gohttp.ListenAndServe(bind.Host, HttpWebsocketHandler(srv, origins))
	}, "http+ws")

	RegisterHandler(func(bind *url.URL, srv *server.Server, opts map[string]any) error {
		tcpAddr, err := net.ResolveTCPAddr("tcp", bind.String())
		if err != nil {
			return err
		}
		tcpListener, err := net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			return err
		}
		for {
			conn, err := tcpListener.Accept()
			if err != nil {
				return err
			}
			go rdwr.NewCodec(conn, conn)
		}
	}, "tcp")
	RegisterDialer(func(ctx context.Context, url string) (jsonrpc.Conn, error) {
		return http.Dial(ctx, http.DefaultH2CClient, strings.Replace(url, "h2c", "http", 1))
	}, "h2c")
	RegisterDialer(func(ctx context.Context, url string) (jsonrpc.Conn, error) {
		return http.Dial(ctx, nil, url)
	}, "https", "http")
	RegisterDialer(func(ctx context.Context, url string) (jsonrpc.Conn, error) {
		return websocket.DialWebsocket(ctx, url, "")
	}, "wss", "ws")

	RegisterDialer(func(ctx context.Context, url string) (jsonrpc.Conn, error) {
		tcpAddr, err := net.ResolveTCPAddr("tcp", url)
		if err != nil {
			return nil, err
		}
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			return nil, err
		}
		return rdwr.NewClient(conn, conn), nil
	}, "tcp")
}
