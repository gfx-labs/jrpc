package codecs

import (
	"gfx.cafe/open/jrpc/contrib/codecs/http"
	"gfx.cafe/open/jrpc/contrib/codecs/inproc"
	"gfx.cafe/open/jrpc/contrib/codecs/websocket"
)

var NewInProc = inproc.NewCodec
var WebsocketHandler = websocket.WebsocketHandler
var HttpHandler = http.HttpHandler
