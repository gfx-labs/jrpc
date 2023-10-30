package server

import (
	"context"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

type peerInfoContextKey struct{}

// PeerInfoFromContext returns information about the client's network connection.
// Use this with the context passed to RPC method handler functions.
//
// The zero value is returned if no connection info is present in ctx.
func PeerInfoFromContext(ctx context.Context) jsonrpc.PeerInfo {
	info, _ := ctx.Value(peerInfoContextKey{}).(jsonrpc.PeerInfo)
	return info
}
func ContextWithPeerInfo(ctx context.Context, c jsonrpc.PeerInfo) context.Context {
	return context.WithValue(ctx, peerInfoContextKey{}, c)
}
