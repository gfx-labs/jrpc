# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

JRPC is a bottom-up JSON-RPC 2.0 implementation primarily designed for hosting Ethereum-like JSON-RPC requests. It follows HTTP-style request/response semantics similar to Go's `net/http` package.

## Architecture

The codebase follows a modular architecture with clear separation between core functionality and optional extensions:

- **Core (`pkg/`)**: Protocol implementation, server, and client utilities
  - `jsonrpc/`: Core types (Request, Response, Handler, ResponseWriter)
  - `server/`: Server implementation using codec abstractions
  - `clientutil/`: Client-side utilities
  - `jrpctest/`: Comprehensive test suite for protocol compliance

- **Extensions (`contrib/`)**:
  - `codecs/`: Transport implementations (HTTP, WebSocket, Reader/Writer)
  - `jmux/`: Chi-style router for method routing
  - `handlers/argreflect/`: Struct reflection for method mounting
  - `middleware/`: Pre-built middleware (logging, recovery, timeout)
  - `extension/subscription/`: WebSocket subscription support

- **Examples (`example/`)**: Working examples demonstrating usage patterns

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

1. **Handler Interface**: Similar to `http.Handler`, methods implement:
   ```go
   func(w jsonrpc.ResponseWriter, r *jsonrpc.Request)
   ```

2. **Middleware Pattern**: Extensions are implemented as middleware wrapping handlers

3. **Codec Abstraction**: Transport-agnostic design using `codec.ReaderWriter` interface

4. **Router Pattern**: Use `jmux.Router` for method routing similar to HTTP routers

## Testing Approach

- Protocol compliance tests in `pkg/jrpctest/`
- Test fixtures for JSON-RPC edge cases in `pkg/jrpctest/testdata/`
- Transport-specific test makers for different codecs
- Table-driven tests using `testify` assertions
- Benchmark suite for performance validation

## CI/CD Pipeline

GitLab CI runs three stages:
1. **test**: Executes tests with race detection
2. **lint**: Runs golangci-lint with project configuration
3. **coverage**: Generates coverage reports (text, XML, Cobertura)