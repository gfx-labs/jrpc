## jrpc

this is a bottom up implementation of jsonrpc2, primarily made for hosting eth-like jsonrpc requests.

packages which are incomplete are marked as such (WIP)

```
exports.go       - export things in subpackages to jrpc namespace, cleaning up the public use package
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
  jmux/            - a chi based router which satisfies the jrpc.Handler interface
  handlers/        - special jrpc handlers
    argreflect/      - go-ethereum style struct reflection
  middleware/      - pre implemented middleware
  openrpc/         - openapi specification implementation
  subscription/    - WIP: subscription engine for go-ethereum style subs

```

