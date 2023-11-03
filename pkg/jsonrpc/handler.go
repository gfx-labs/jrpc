package jsonrpc

// http.handler, but for jrpc
type Handler interface {
	ServeRPC(w ResponseWriter, r *Request)
}

// type check for handlerfunc
var _ Handler = (HandlerFunc)(nil)

// http.HandlerFunc,but for jrpc
type HandlerFunc func(w ResponseWriter, r *Request)

// ServeRPC implements (jsonrpc.Handler).ServeRPC
func (fn HandlerFunc) ServeRPC(w ResponseWriter, r *Request) {
	(fn)(w, r)
}
