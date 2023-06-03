## jrpc

this is a bottom up implementation of jsonrpc2, primarily made for hosting eth-like jsonrpc requests.

many packages are unused / incomplete. the one which are done i have put below in the readme

```
exports.go       - export things in subpackages to jrpc namespace, cleaning up the public use package
pkg/             - packages for implementing jrpc
  clientutil/      - common utilities for client implementations to use
    idreply.go       - generalizes making a request with an incrementing id, then waiting on it
    helper.go        - helpers for decoding messages, etc
  codec/           - codec related things. used by client and server implementations
    errors.go        - jsonrpc2 error codes and marshaling
    json.go          - jsonrpc2 json rules, encoding, decoding
    peer.go          - peerinfo
    transport.go     - define ReaderWriter interface
    wire.go          - jsonrpc2 wire protocol marshaling, like ID and Version
    server.go        - a server server implementation that uses the codec
    jrpc.go          - define the Handler, HandlerFunc, and ResponseWriter
    reqresp.go       - define Request, Response, along with json marshaling for the request
  jrpctest/        - utilities for testing client and server.
    suite.go         - implementors of client and server should pass this
contrib/         - packages that add to jrpc
  jmux/            - a chi based router which satisfies the jrpc.Handler interface
  handlers/        - special jrpc handlers
    argreflect/      - go-ethereum style struct reflection
  middleware/      - pre implemented middleware
  openrpc/         - openapi specification implementation
  subscription     - WIP: subscription engine for go-ethereum style subs


```
