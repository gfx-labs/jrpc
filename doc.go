// Package jrpc provides a bottom-up JSON-RPC 2.0 implementation with HTTP-style
// request/response semantics, designed primarily for hosting Ethereum-like JSON-RPC
// requests but suitable for any JSON-RPC 2.0 application.
//
// The package follows familiar Go patterns, modeling its API after net/http with
// Handler interfaces, ResponseWriter, and Request types. It supports multiple
// transports including HTTP, WebSocket, and raw IO streams, with a pluggable
// codec system for adding new transports.
//
// # Basic Usage
//
// Create a simple echo server:
//
//	r := jmux.NewRouter()
//	r.HandleFunc("echo", func(w jrpc.ResponseWriter, r *jrpc.Request) {
//	    w.Send(r.Params, nil)
//	})
//	http.ListenAndServe(":8080", codecs.HttpWebsocketHandler(r, []string{"*"}))
//
// Make client requests:
//
//	conn, _ := jrpc.Dial("ws://localhost:8080")
//	defer conn.Close()
//
//	result, err := jrpc.Do[string](ctx, conn, "echo", "hello")
//
// # Core Components
//
// Handler Interface: The foundation of the server-side API, similar to http.Handler:
//
//	type Handler interface {
//	    ServeRPC(w ResponseWriter, r *Request)
//	}
//
// ResponseWriter: Used to send responses back to clients:
//
//	type ResponseWriter interface {
//	    Send(v any, err error) error      // Send response
//	    Notify(method string, v any) error // Send notification
//	    ExtraFields() ExtraFields          // Access extra fields
//	}
//
// Request: Contains the JSON-RPC request data:
//
//	type Request struct {
//	    ID     *ID              // Request ID (nil for notifications)
//	    Method string           // Method name
//	    Params json.RawMessage  // Raw parameters
//	    Peer   PeerInfo        // Connection information
//	}
//
// # Features
//
// Protocol Support:
//   - Full JSON-RPC 2.0 specification compliance
//   - Batch request processing with configurable parallelism
//   - Notifications (requests without IDs)
//   - Proper error codes (-32700 to -32603)
//   - Extra fields support for protocol extensions
//
// Transport Flexibility:
//   - HTTP with GET/POST/SSE support
//   - WebSocket for bidirectional communication
//   - Raw IO streams (TCP, Unix sockets, etc.)
//   - In-process communication
//   - Pluggable codec system for custom transports
//
// Router and Middleware:
//   - Chi-style method routing with pattern matching
//   - Middleware chain support for cross-cutting concerns
//   - Method pattern matching and routing
//   - Context propagation throughout the stack
//
// Performance:
//   - Efficient JSON handling with go-faster/jx
//   - Buffer pooling to reduce allocations
//   - Streaming response support
//   - Configurable parallel batch processing
//
// # Error Handling
//
// The package provides typed errors following JSON-RPC 2.0 specification:
//
//	w.Send(nil, jsonrpc.InvalidParams.WithData("missing required field"))
//	w.Send(nil, jsonrpc.MethodNotFound)
//	w.Send(nil, jsonrpc.InternalError.WithData(err.Error()))
//
// Custom errors can implement the Error or DataError interfaces for
// proper JSON-RPC error responses.
//
package jrpc
