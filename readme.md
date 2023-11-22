## jrpc

```go get gfx.cafe/open/jrpc```

this is a bottom up implementation of jsonrpc2, primarily made for hosting eth-like jsonrpc requests.

we extend the eth-rpc reflect based handler with go-http style request/response.

we also make things like subscriptions additional extensions, so they are no longer baked into the rpc package.

most users should mostly access the `jrpc` packages, along with a variety of things in `contrib`

see examples in `examples` folder for usage

it is currently being used in the oku.trade api in proxy, client, and server applications.

## features

 - full jsonrpc2 protocol
 - batch requests + notifications
 - http.Request/http.ResponseWriter style semantics
 - simple but powerful middleware framework
 - subscription framework used by go-ethereum/rpc is implemented as middleware.
 - http (with rest-like access via RPC verb), websocket, io.Reader/io.Writer (tcp, any net.Conn, etc), inproc codecs.
 - using faster json packages (jsoniter, jx)
 - extensions, which allow setting arbitrary fields on the parent object, like in sourcegraph jsonrpc2
 - jmux, which allows for http-like routing, implemented like `go-chi/v5`, except for jsonrpc2 paths
 - argreflect, which allows mounting methods on structs to the rpc engine, like go-ethereum/rpc
 - openrpc schema parser and code generator


## maybe outdated but somewhat useful contribution info

basic structure

```
exports.go       - export things in subpackages to jrpc namespace, cleaning up the public use package.
pkg/             - packages for implementing jrpc
  clientutil/      - common utilities for client implementations to use
    idreply.go       - generalizes making a request with an incrementing id, then waiting on it
    helper.go        - helpers for decoding messages, etc
  codec/           - codec related things. to implement new codecs, use this package
    errors.go        - jsonrpc2 error codes and marshaling
    json.go          - jsonrpc2 json rules, encoding, decoding
    peer.go          - peerinfo
    transport.go     - define ReaderWriter interface
    wire.go          - jsonrpc2 wire protocol marshaling, like ID and Version
    jrpc.go          - define the Handler, HandlerFunc, and ResponseWriter
    reqresp.go       - define Request, Response, along with json marshaling for the request
  server/            - server implementation
    server.go        - a simple server implementation that uses a codec.ReaderWriter
  jrpctest/        - utilities for testing client and server.
    suite.go         - implementors of client and server should pass this
contrib/         - packages that add to jrpc
  codecs/          - client and server transport implementations
    codecs.go        - dialers for all finished codecs
    http/              - http based codec
      codec_test.go      - general tests that all must pass
      client.go          - codec.Conn implementation
      codec.go           - codec.ReaderWriter implementaiton
      const.go           - constants
      handler.go         - http handler
      http_test.go       - http specific tests
    websocket/         - websocket basec codec
      codec_test.go      - general tests that all must pass
      client.go          - codec.Conn implementation
      codec.go           - codec.ReadWriter implementation
      const.go           - constants
      dial.go            - websocket dialer
      handler.go         - http handler
      websocket_test.go  - websocket specific tests
    rdwr/              - rdwr based codec. can be used to implement other codecs
    inproc/            - inproc based codec
  openrpc/         - openapi specification implementation
  jmux/            - a chi based router which satisfies the jrpc.Handler interface
  handlers/        - special jrpc handlers
    argreflect/      - go-ethereum style struct reflection
  middleware/      - pre implemented middleware
  extension/       - extensions to the protocol
    subscription/    - WIP: subscription engine for go-ethereum style subs

```

