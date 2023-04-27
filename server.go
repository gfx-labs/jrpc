package jrpc

import (
	"context"
	"sync/atomic"

	"gfx.cafe/open/jrpc/codec"
	mapset "github.com/deckarep/golang-set"
)

// Server is an RPC server.
type Server struct {
	services Handler
	run      int32
	codecs   mapset.Set
}

// NewServer creates a new server instance with no registered handlers.
func NewServer(r Handler) *Server {
	server := &Server{
		codecs: mapset.NewSet(),
		run:    1,
	}
	server.services = r
	// Register the default service providing meta information about the RPC service such
	// as the services and methods it offers.
	return server
}

// ServeCodec reads incoming requests from codec, calls the appropriate callback and writes
// the response back using the given codec. It will block until the codec is closed or the
// server is stopped. In either case the codec is closed.
func (s *Server) ServeCodec(codec codec.ReaderWriter) {
	defer codec.Close()

	// Don't serve if server is stopped.
	if atomic.LoadInt32(&s.run) == 0 {
		return
	}
	// Add the codec to the set so it can be closed by Stop.
	s.codecs.Add(codec)
	defer s.codecs.Remove(codec)

	// TODO: handle this
	// c := initClient(codec, s.services)
	// <-codec.Closed()
	// c.Close()
}

// Stop stops reading new requests, waits for stopPendingRequestTimeout to allow pending
// requests to finish, then closes all codecs which will cancel pending requests and
// subscriptions.
func (s *Server) Stop() {
	if atomic.CompareAndSwapInt32(&s.run, 1, 0) {
		s.codecs.Each(func(c any) bool {
			c.(codec.ReaderWriter).Close()
			return true
		})
	}
}

type peerInfoContextKey struct{}

// PeerInfoFromContext returns information about the client's network connection.
// Use this with the context passed to RPC method handler functions.
//
// The zero value is returned if no connection info is present in ctx.
func PeerInfoFromContext(ctx context.Context) codec.PeerInfo {
	info, _ := ctx.Value(peerInfoContextKey{}).(codec.PeerInfo)
	return info
}
