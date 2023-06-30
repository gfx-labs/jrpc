package codecs

import (
	"gfx.cafe/open/jrpc/contrib/codecs/http"
	"gfx.cafe/open/jrpc/contrib/codecs/inproc"
	"gfx.cafe/open/jrpc/contrib/codecs/websocket"
	"gfx.cafe/open/jrpc/pkg/server"

	gohttp "net/http"
)

var NewInProc = inproc.NewCodec
var WebsocketHandler = websocket.WebsocketHandler
var HttpHandler = http.HttpHandler

var HttpWebsocketHandler = func(srv *server.Server, origins []string) gohttp.Handler {
	cwss := WebsocketHandler(srv, origins)
	chttp := HttpHandler(srv)
	return gohttp.HandlerFunc(func(w gohttp.ResponseWriter, r *gohttp.Request) {
		if r.Header.Get("upgrade") != "" {
			cwss.ServeHTTP(w, r)
			return
		}
		chttp.ServeHTTP(w, r)
	})
}
