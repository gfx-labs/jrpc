package jrpc

import (
	"net/http"
)

// http.handler, but for jrpc
type Handler interface {
	ServeRPC(w ResponseWriter, r *Request)
}

// http.HandlerFunc,but for jrpc
type HandlerFunc func(w ResponseWriter, r *Request)

func (fn HandlerFunc) ServeRPC(w ResponseWriter, r *Request) {
	(fn)(w, r)
}

// http.ResponseWriter interface, but for jrpc
type ResponseWriter interface {
	Send(v any, err error) error
	Option(k string, v any)
	Header() http.Header

	Notify(v any) error
}
