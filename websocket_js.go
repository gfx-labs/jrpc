package jrpc

import (
	"context"

	"nhooyr.io/websocket"
)

// that is listening on the given endpoint using the provided dialer.
func DialWebsocketWithDialer(ctx context.Context, endpoint, origin string, opts *websocket.DialOptions) (*Client, error) {
	endpoint, _, err := wsClientHeaders(endpoint, origin)
	if err != nil {
		return nil, err
	}
	return newClient(ctx, func(cctx context.Context) (ServerCodec, error) {
		conn, resp, err := websocket.Dial(cctx, endpoint, opts)
		if err != nil {
			hErr := wsHandshakeError{err: err}
			if resp != nil {
				hErr.status = resp.Status
			}
			return nil, hErr
		}
		out := newWebsocketCodec(resp.Request.Context(), conn, endpoint, nil)
		return out, err
	})
}

// DialWebsocket creates a new RPC client that communicates with a JSON-RPC server
// that is listening on the given endpoint.
//
// The context is used for the initial connection establishment. It does not
// affect subsequent interactions with the client.
func DialWebsocket(ctx context.Context, endpoint, origin string) (*Client, error) {
	endpoint, _, err := wsClientHeaders(endpoint, origin)
	if err != nil {
		return nil, err
	}
	dialer := &websocket.DialOptions{}
	return DialWebsocketWithDialer(ctx, endpoint, origin, dialer)
}
