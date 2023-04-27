package codec

import "net/http"

// PeerInfo contains information about the remote end of the network connection.
//
// This is available within RPC method handlers through the context. Call
// PeerInfoFromContext to get information about the client connection related to
// the current method call.
type PeerInfo struct {
	// Transport is name of the protocol used by the client.
	// This can be "http", "ws" or "ipc".
	Transport string

	// Address of client. This will usually contain the IP address and port.
	RemoteAddr string

	// Addditional information for HTTP and WebSocket connections.
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
