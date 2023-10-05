package server

import (
	"net/http"

	"gfx.cafe/open/jrpc/pkg/codec"
)

var _ codec.ResponseWriter = (*callRespWriter)(nil)

type callRespWriter struct {
	msg *codec.Message

	pkt *codec.Message

	dat    any
	skip   bool
	header http.Header

	notifications func(env *notifyEnv) error
}

func (c *callRespWriter) Send(v any, err error) error {
	if err != nil {
		c.pkt.Error = err
		return nil
	}
	c.dat = v
	return nil
}

func (c *callRespWriter) SetExtraField(k string, v any) error {
	c.pkt.SetExtraField(k, v)
	return nil
}

func (c *callRespWriter) Header() http.Header {
	return c.header
}

func (c *callRespWriter) Notify(method string, v any) error {
	return c.notifications(&notifyEnv{
		method: method,
		dat:    v,
		extra:  c.pkt.ExtraFields,
	})
}
