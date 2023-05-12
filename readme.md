## jrpc

this is a bottom up implementation of jsonrpc2, primarily made for hosting eth-like jsonrpc requests.

many packages are unused / incomplete. the one which are done i have put below in the readme

```
conn.go        -  defines the interface a json rpc client
jrpc.go        - define the Handler, HandlerFunc, and ResponseWriter
request.go     - define Request, along with json marshaling for the request
response.go    - define Response, along with json marshaling for the response
server.go      - implemntation of a server, which uses a backing codec.ReaderWriter
codec/         - codec related things
  stdio/       - implementation of jrpc.Conn and codec.ReaderWriter
  errors.go    - jsonrpc2 error codes and marshaling
  json.go      - jsonrpc2 json rules, encoding, decoding
  peer.go      - peerinfo
  transport.go - define ReaderWriter interface
  wire.go      - jsonrpc2 wire protocol marshaling, like ID and Version
jmux/          - a chi based router which satisfies the jrpc.Handler interface
clientutil/    - common utilities for client implementations to use
  idreply.go   - generalizes making a request with an incrementing id, then waiting on it
```
