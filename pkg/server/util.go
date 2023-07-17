package server

import (
	"context"

	"gfx.cafe/open/jrpc/pkg/codec"
)

type peerInfoContextKey struct{}

// PeerInfoFromContext returns information about the client's network connection.
// Use this with the context passed to RPC method handler functions.
//
// The zero value is returned if no connection info is present in ctx.
func PeerInfoFromContext(ctx context.Context) codec.PeerInfo {
	info, _ := ctx.Value(peerInfoContextKey{}).(codec.PeerInfo)
	return info
}
func ContextWithPeerInfo(ctx context.Context, c codec.PeerInfo) context.Context {
	return context.WithValue(ctx, peerInfoContextKey{}, c)
}
