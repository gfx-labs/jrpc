package jrpc

import (
	"context"
	"io"
	"net/http"
	"sync/atomic"

	mapset "github.com/deckarep/golang-set"
	"tuxpa.in/a/zlog/log"
)

const (
	MetadataApi = "rpc"
	EngineApi   = "engine"
)

// CodecOption specifies which type of messages a codec supports.

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
func (s *Server) ServeCodec(codec ServerCodec) {
	defer codec.Close()

	// Don't serve if server is stopped.
	if atomic.LoadInt32(&s.run) == 0 {
		return
	}

	// Add the codec to the set so it can be closed by Stop.
	s.codecs.Add(codec)
	defer s.codecs.Remove(codec)

	c := initClient(codec, s.services)
	<-codec.Closed()
	c.Close()
}

// serveSingleRequest reads and processes a single RPC request from the given codec. This
// is used to serve HTTP connections. Subscriptions and reverse calls are not allowed in
// this mode.
func (s *Server) serveSingleRequest(ctx context.Context, codec ServerCodec) {
	// Don't serve if server is stopped.
	if atomic.LoadInt32(&s.run) == 0 {
		return
	}
	// create a new handler for this context
	h := newHandler(ctx, codec, s.services)
	defer h.close(io.EOF, nil)

	// read the HTTP body
	reqs, batch, err := codec.ReadBatch()
	if err != nil {
		if err != io.EOF {
			codec.WriteJSON(ctx, errorMessage(&invalidMessageError{"parse error"}))
		}
		return
	}
	if batch {
		h.handleBatch(reqs)
	} else {
		h.handleMsg(reqs[0])
	}
}

// Stop stops reading new requests, waits for stopPendingRequestTimeout to allow pending
// requests to finish, then closes all codecs which will cancel pending requests and
// subscriptions.
func (s *Server) Stop() {
	if atomic.CompareAndSwapInt32(&s.run, 1, 0) {
		log.Debug().Msg("RPC server shutting down")
		s.codecs.Each(func(c any) bool {
			c.(ServerCodec).Close()
			return true
		})
	}
}

// Deprecated: RPCService gives meta information about the server.
// e.g. gives information about the loaded modules.
type RPCService struct {
	server *Server
}

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

	WriteHeaders http.Header
}

type peerInfoContextKey struct{}

// PeerInfoFromContext returns information about the client's network connection.
// Use this with the context passed to RPC method handler functions.
//
// The zero value is returned if no connection info is present in ctx.
func PeerInfoFromContext(ctx context.Context) PeerInfo {
	info, _ := ctx.Value(peerInfoContextKey{}).(PeerInfo)
	return info
}
