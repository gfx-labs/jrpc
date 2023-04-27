package websocket

import "time"

const (
	MaxRequestContentLength = 1024 * 1024 * 5
	ContentType             = "application/json"
	WsReadBuffer            = 1024
	WsWriteBuffer           = 1024
	WsPingInterval          = 60 * time.Second
	WsPingWriteTimeout      = 5 * time.Second
	WsPongTimeout           = 30 * time.Second
	WsMessageSizeLimit      = 128 * 1024 * 1024
)
