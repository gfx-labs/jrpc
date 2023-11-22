package jsonrpc

import (
	"net/http"
)

type PeerInfo struct {
	// Transport is name of the protocol used by the client.
	Transport string

	// Address of client. This will usually contain the IP address and port.
	RemoteAddr string

	// Additional information for HTTP and WebSocket connections.
	HTTP HttpInfo
}

type HttpInfo struct {
	// Protocol version, i.e. "HTTP/1.1". This is not set for WebSocket.
	Version string
	// Header values sent by the client.
	UserAgent string
	Origin    string
	Host      string

	Headers http.Header
}
