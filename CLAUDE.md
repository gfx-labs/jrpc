# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

JRPC is a bottom-up JSON-RPC 2.0 implementation primarily designed for hosting Ethereum-like JSON-RPC requests. It follows HTTP-style request/response semantics similar to Go's `net/http` package.

Module: `gfx.cafe/open/jrpc` (requires Go 1.21+)

## Architecture

The codebase follows a modular architecture with clear separation between core functionality and optional extensions:

### Core (`pkg/`) - Protocol Implementation

- **`jsonrpc/`**: Core types and interfaces
  - `Handler` interface: `ServeRPC(w ResponseWriter, r *Request)`
  - `Request` & `ResponseWriter`: Core request/response types
  - `Message` types: For streaming JSON-RPC messages
  - Error types: Full JSON-RPC 2.0 error specification (-32700 to -32603)
  - Context-aware design throughout

- **`server/`**: Transport-agnostic server implementation
  - Uses `ReaderWriter` codec abstraction
  - Supports single and batch request processing
  - Configurable parallel batch processing
  - Stream-based response writing with proper locking

- **`clientutil/`**: Client-side utilities
  - Helper functions for RPC calls
  - ID management for request/response correlation

- **`jjson/`**: JSON utilities with buffer pooling

### Internal (`internal/`)

- **`jrpctest/`**: Internal test framework (not exposed to external packages)
  - Protocol compliance tests
  - Test fixtures in `testdata/`
  - Server/client test makers
  - Transport-specific test utilities

### Extensions (`contrib/`)

- **`codecs/`**: Transport implementations
  - HTTP: GET/POST/SSE support with proper CORS handling
  - WebSocket: Bidirectional communication
  - Reader/Writer: For testing and custom transports
  - Registry pattern for codec discovery

- **`jmux/`**: Chi-style method router
  - Pattern-based method routing
  - Middleware chain support
  - Method routing patterns
  - Tree-based routing algorithm

- **`handlers/argreflect/`**: Reflection-based handler mounting

- **`middleware/`**: Pre-built middleware
  - Recoverer: Panic recovery with stack traces
  - Logger: Request/response logging
  - Timeout: Context-based request timeouts
  - RequestID: Request correlation
  - HTTP-specific: CORS, headers, etc.

- **`extension/subscription/`**: WebSocket subscription support
  - Pub/sub pattern implementation
  - Context-based notifier system
  - Automatic cleanup on disconnect

### Examples (`example/`)
- `echo/`: Basic echo server
- `jmux/`: Router usage examples
- `stream/`: Streaming responses
- `subscription/`: WebSocket subscriptions

## Development Commands

```bash
# Install dependencies
go mod download

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector (CI default)
go test -race ./...

# Run benchmarks
go test -bench=. ./benchmark/

# Build all packages
go build ./...

# Run a specific test
go test -v ./pkg/jsonrpc -run TestRequestParsing

# Run echo server example
go run example/echo/main.go

# Run linter (requires golangci-lint)
golangci-lint run
```

## Key Design Patterns

### 1. Handler Interface
Similar to `http.Handler`, methods implement:
```go
type Handler interface {
    ServeRPC(w ResponseWriter, r *Request)
}
```

### 2. Codec Abstraction
Transport-agnostic design using:
```go
type ReaderWriter interface {
    ReadBatch(ctx context.Context) (Bundle, error)
    Write([]byte) (int, error)
    PeerInfo() PeerInfo
    Close() error
}
```

### 3. Middleware Pattern
Standard middleware chaining:
```go
type Middleware func(Handler) Handler
```

### 4. Router Pattern
Use `jmux.Router` for method routing:
```go
r := jmux.NewRouter()
r.HandleFunc("method", handler)
r.Mount("/prefix", subrouter)
```

### 5. Error Handling
JSON-RPC 2.0 compliant error types:
```go
type Error interface {
    Error() string
    ErrorCode() int
}

type DataError interface {
    Error
    ErrorData() any
}
```

## Testing Approach

- Protocol compliance tests in `internal/jrpctest/`
- Test fixtures for JSON-RPC edge cases in `internal/jrpctest/testdata/`
- Transport-specific test makers for different codecs
- Table-driven tests using `testify` assertions
- Benchmark suite for performance validation

## CI/CD Pipeline

GitLab CI runs three stages:
1. **test**: Executes tests with race detection
2. **lint**: Runs golangci-lint with project configuration
3. **coverage**: Generates coverage reports (text, XML, Cobertura)

## Common Usage Patterns

### Basic Server Setup
```go
package main

import (
    "net/http"
    "gfx.cafe/open/jrpc/contrib/codecs"
    "gfx.cafe/open/jrpc/contrib/jmux"
    "gfx.cafe/open/jrpc/pkg/jsonrpc"
)

func main() {
    r := jmux.NewRouter()
    
    // Simple handler function
    r.HandleFunc("echo", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
        w.Send(r.Params, nil)
    })
    
    // With middleware
    r.Use(middleware.Recoverer)
    r.Use(middleware.Logger)
    
    // Serve over HTTP and WebSocket
    http.ListenAndServe(":8080", codecs.HttpWebsocketHandler(r, []string{"*"}))
}
```


### Client Usage
```go
import "gfx.cafe/open/jrpc"

// Dial connection
conn, err := jrpc.Dial("ws://localhost:8080")
if err != nil {
    panic(err)
}
defer conn.Close()

// Make requests using generic helpers
result, err := jrpc.Do[string](ctx, conn, "echo", "hello")

// Or use conn.Do() directly
var blockNumber string
err = conn.Do(ctx, &blockNumber, "eth_blockNumber", nil)
```

## Extension Points

To extend the codebase:

1. **New Transport**: Implement `ReaderWriter` interface in `contrib/codecs/`
2. **Custom Middleware**: Add to `contrib/middleware/` following the middleware pattern
3. **Router Features**: Extend `contrib/jmux/` with new routing capabilities
4. **Protocol Extensions**: Add to `contrib/extension/` (e.g., subscriptions)
5. **Custom Error Types**: Implement `Error` or `DataError` interfaces

## Performance Considerations

- Uses `go-faster/jx` for efficient JSON encoding
- Buffer pooling via `jjson` package to reduce allocations
- Configurable parallel batch processing
- Stream-based response writing for large payloads
- Proper resource cleanup and context cancellation

## Best Practices

1. **Always use context** for cancellation and timeouts
2. **Use method routing** for organizing large APIs
3. **Use middleware** for cross-cutting concerns (logging, auth, etc.)
4. **Leverage codec registry** for transport flexibility
5. **Follow test patterns** in `internal/jrpctest` for new features
6. **Handle errors properly** using the built-in error types
7. **Use streaming APIs** for efficient batch processing
8. **Close resources** properly (codecs, servers, etc.)