package websocket

import (
	"context"

	"nhooyr.io/websocket"
)

// DialWebsocket creates a new RPC client that communicates with a JSON-RPC server
// that is listening on the given endpoint.
//
// The context is used for the initial connection establishment. It does not
// affect subsequent interactions with the client.
func DialWebsocket(ctx context.Context, endpoint, origin string) (*Client, error) {
	endpoint, header, err := WsClientHeaders(endpoint, origin)
	if err != nil {
		return nil, err
	}
	dialer := &websocket.DialOptions{
		CompressionMode:      websocket.CompressionContextTakeover,
		CompressionThreshold: 4096,
		HTTPHeader:           header,
	}
	return DialWebsocketWithDialer(ctx, endpoint, origin, dialer)
}

// that is listening on the given endpoint using the provided dialer.
func DialWebsocketWithDialer(ctx context.Context, endpoint, origin string, opts *websocket.DialOptions) (*Client, error) {
	endpoint, header, err := WsClientHeaders(endpoint, origin)
	if err != nil {
		return nil, err
	}
	opts.HTTPHeader = header
	return newClient(ctx, func(cctx context.Context) (*websocket.Conn, error) {
		conn, resp, err := websocket.Dial(cctx, endpoint, opts)
		if err != nil {
			hErr := WsHandshakeError{err: err}
			if resp != nil {
				hErr.status = resp.Status
			}
			return nil, hErr
		}
		return conn, err
	})
}
