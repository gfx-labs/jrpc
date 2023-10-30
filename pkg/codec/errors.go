package codec

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-faster/jx"
)

// Error types defined below are the built-in JSON-RPC errors.

var (
	_ Error = new(ErrorMethodNotFound)
	_ Error = new(ErrorSubscriptionNotFound)
	_ Error = new(ErrorParse)
	_ Error = new(ErrorInvalidRequest)
	_ Error = new(ErrorInvalidMessage)
	_ Error = new(ErrorInvalidParams)
)

const (
	ErrorCodeDefault     = -32000
	ErrorCodeApplication = -32080
	ErrorCodeJrpc        = -42000
)

var (
	ErrIllegalExtraField    = errors.New("invalid extra field")
	ErrSendAlreadyCalled    = errors.New("send already called")
	ErrCantSendNotification = errors.New("can't send to a notification")
)

// Error wraps RPC errors, which contain an error code in addition to the message.
type Error interface {
	Error() string  // returns the message
	ErrorCode() int // returns the code
}

// A DataError contains some data in addition to the error message.
type DataError interface {
	Error() string  // returns the message
	ErrorCode() int // returns the error code
	ErrorData() any // returns the error data
}

func EncodeError(enc *jx.Encoder, err error) error {
	enc.Obj(func(e *jx.Encoder) {
		switch er := err.(type) {
		case DataError:
			e.Field("code", func(e *jx.Encoder) { e.Int(er.ErrorCode()) })
			e.Field("message", func(e *jx.Encoder) { e.Str(er.Error()) })
			if dat := er.ErrorData(); dat != nil {
				data, err := json.Marshal(er.ErrorData())
				if err != nil {
					data = []byte(`"failed to marshal error data"`)
				}
				e.Field("data", func(e *jx.Encoder) {
					e.Raw(data)
				})
			}
		case Error:
			e.FieldStart("code")
			e.Int(er.ErrorCode())
			e.FieldStart("message")
			e.Str(er.Error())
		default:
			e.Field("code", func(e *jx.Encoder) { e.Int(-32000) })
			e.Field("message", func(e *jx.Encoder) { e.Str(er.Error()) })
		}
	})
	return nil
}

func WrapErr(data any, code int, err error) error {
	return &jrpcErr{
		data: data,
		err:  err,
		code: code,
	}
}

type jrpcErr struct {
	data any
	err  error
	code int
}

func (j *jrpcErr) ErrorData() any {
	return j.data
}

func (j *jrpcErr) Error() string {
	return j.err.Error()
}

func (j *jrpcErr) ErrorCode() int {
	return j.code
}

type ErrorMethodNotFound struct{ method string }

func (e *ErrorMethodNotFound) ErrorCode() int { return -32601 }
func (e *ErrorMethodNotFound) Error() string {
	return fmt.Sprintf("the method %s does not exist/is not available", e.method)
}

func NewMethodNotFoundError(method string) *ErrorMethodNotFound {
	return &ErrorMethodNotFound{
		method: method,
	}
}

type ErrorSubscriptionNotFound struct{ namespace, subscription string }

func (e *ErrorSubscriptionNotFound) ErrorCode() int { return -32601 }
func (e *ErrorSubscriptionNotFound) Error() string {
	return fmt.Sprintf("no %q subscription in %s namespace", e.subscription, e.namespace)
}

// Invalid JSON was received by the server.
type ErrorParse struct{ message string }

func (e *ErrorParse) ErrorCode() int { return -32700 }
func (e *ErrorParse) Error() string  { return e.message }

// received message isn't a valid request
type ErrorInvalidRequest struct{ message string }

func NewInvalidRequestError(message string) *ErrorInvalidRequest {
	return &ErrorInvalidRequest{message: message}
}

func (e *ErrorInvalidRequest) ErrorCode() int { return -32600 }
func (e *ErrorInvalidRequest) Error() string  { return e.message }

// received message is invalid
type ErrorInvalidMessage struct{ message string }

func (e *ErrorInvalidMessage) ErrorCode() int { return -32700 }
func (e *ErrorInvalidMessage) Error() string  { return e.message }

// unable to decode supplied params, or an invalid number of parameters
type ErrorInvalidParams struct{ message string }

func NewInvalidParamsError(message string) *ErrorInvalidParams {
	return &ErrorInvalidParams{message: message}
}
func (e *ErrorInvalidParams) ErrorCode() int { return -32602 }
func (e *ErrorInvalidParams) Error() string  { return e.message }

// unable to decode supplied params, or an invalid number of parameters
type ErrorInternalError struct{ message string }

func NewInternalError(message string) *ErrorInternalError {
	return &ErrorInternalError{message: message}
}
func (e *ErrorInternalError) ErrorCode() int { return -32603 }
func (e *ErrorInternalError) Error() string  { return e.message }

// HTTPError is returned by client operations when the HTTP status code of the
// response is not a 2xx status.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (err HTTPError) Error() string {
	if len(err.Body) == 0 {
		return err.Status
	}
	return fmt.Sprintf("%v: %s", err.Status, err.Body)
}
