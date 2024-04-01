package jsonrpc

import (
	"fmt"
)

type ResponseController struct {
	rw ResponseWriter
}

func NewResponseController(rw ResponseWriter) *ResponseController {
	return &ResponseController{rw}
}

type rwUnwrapper interface {
	Unwrap() ResponseWriter
}

func (c *ResponseController) Hijack() (send MessageStreamer, notify MessageStreamer, err error) {
	rw := c.rw
	for {
		switch t := rw.(type) {
		case Hijacker:
			return t.Hijack()
		case rwUnwrapper:
			rw = t.Unwrap()
		default:
			return nil, nil, errNotSupported()
		}
	}
}

func errNotSupported() error {
	return fmt.Errorf("%w", ErrNotSupported)
}
