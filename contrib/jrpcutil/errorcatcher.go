package jrpcutil

import "github.com/gfx-labs/jrpc/pkg/jsonrpc"

type ErrorRecorder struct {
	jsonrpc.ResponseWriter

	err error
}

func (e *ErrorRecorder) Send(v any, err error) error {
	newErr := e.ResponseWriter.Send(v, err)
	if err != nil {
		e.err = err
	}
	return newErr
}

func (e *ErrorRecorder) Error() error {
	return e.err
}
