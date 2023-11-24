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
	HTTP *http.Request `json:"-"`
}
