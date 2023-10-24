package codecs

import (
	"fmt"

	"gfx.cafe/open/jrpc/contrib/codecs/http"
	"gfx.cafe/open/jrpc/contrib/codecs/inproc"
	"gfx.cafe/open/jrpc/contrib/codecs/websocket"
	"gfx.cafe/open/jrpc/pkg/server"

	gohttp "net/http"
	"net/url"
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

func ListenAndServe(u string, srv *server.Server, opts map[string]any) error {
	pu, err := url.Parse(u)
	if err != nil {
		return err
	}
	if opts == nil {
		opts = map[string]any{}
	}
	handler := handlerFuncs[pu.Scheme]
	if handler == nil {
		return fmt.Errorf("%w: %s", ErrSchemeNotSupported, pu.Scheme)
	}
	return handler(pu, srv, opts)
}
