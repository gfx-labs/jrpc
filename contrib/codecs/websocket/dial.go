package websocket

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"

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
	conn, _, err := websocket.Dial(ctx, endpoint, opts)
	if err != nil {
		return nil, err
	}
	return newClient(conn)
}

func WsClientHeaders(endpoint, origin string) (string, http.Header, error) {
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return endpoint, nil, err
	}
	header := make(http.Header)
	if origin != "" {
		header.Add("origin", origin)
	}
	if endpointURL.User != nil {
		b64auth := base64.StdEncoding.EncodeToString([]byte(endpointURL.User.String()))
		header.Add("authorization", "Basic "+b64auth)
		endpointURL.User = nil
	}
	return endpointURL.String(), header, nil
}
