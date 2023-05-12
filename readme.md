## jrpc

this is a bottom up implementation of jsonrpc2, primarily made for hosting eth-like jsonrpc requests.

structure:

```
conn.go  -  defines the interface a json rpc client
jrpc.go - define the Handler, HandlerFunc, and ResponseWriter
request.go - define Request, along with json marshaling for the request
response.go - define Response, along with json marshaling for the response
```
