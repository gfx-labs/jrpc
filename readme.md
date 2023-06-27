## jrpc

it is a jsonrpc2-like framework, allowing for json rpc over multiple different protocols.

it also has interfaces similar to chi for routing, middleware, etc.

different transports are defined, such as http, websocket, or ipc, (net.Conn), stdio (io.Reader/io.Writer)

it also supports subscriptions, such as those in the original ethereum json-rpc specification

it is based off of the http router chi, and the go-ethereum, sourcegraph, and gopls json-rpc packages.


