package inproc

import (
	"encoding/json"
	"io"
	"net/http"

	"gfx.cafe/open/jrpc/pkg/codec"
)

type callRespWriter struct {
	msg    *codec.Message
	w      io.Writer
	header http.Header
}

func (c *callRespWriter) Send(v any, err error) error {
	return nil
}

func (c *callRespWriter) Option(k string, v any) {
	// no options for now
}

func (c *callRespWriter) Header() http.Header {
	return c.header
}

func (c *callRespWriter) Notify(v any) error {
	return json.NewEncoder(w).Encode(v)
}
