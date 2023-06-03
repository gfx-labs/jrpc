package codec

import (
	"encoding/json"
	"fmt"

	"github.com/go-faster/jx"
)

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
			data, err := json.Marshal(er.ErrorData())
			if err != nil {
				data = []byte(`"failed to marshal error data"`)
			}
			e.Field("code", func(e *jx.Encoder) { e.Int(er.ErrorCode()) })
			e.Field("message", func(e *jx.Encoder) { e.Str(er.Error()) })
			e.Field("data", func(e *jx.Encoder) {
				e.Raw(data)
			})
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

type JrpcErr struct {
	Data any
}

func (j *JrpcErr) ErrorData() any {
	return j.Data
}

func (j *JrpcErr) Error() string {
	return "Jrpc Error"
}

func (j *JrpcErr) ErrorCode() int {
	return ErrorCodeJrpc
}

func WrapJrpcErr(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %w", &JrpcErr{}, err)
}

func MakeJrpcErr(s string) error {
	return fmt.Errorf("%w: %s", &JrpcErr{}, s)
}

// Error types defined below are the built-in JSON-RPC errors.

var (
	_ Error = new(ErrorMethodNotFound)
	_ Error = new(ErrorSubscriptionNotFound)
	_ Error = new(ErrorParse)
	_ Error = new(ErrorInvalidRequest)
	_ Error = new(ErrorInvalidMessage)
	_ Error = new(ErrorInvalidParams)
)

const ErrorCodeDefault = -32000

const ErrorCodeApplication = -32080

const ErrorCodeJrpc = -42000

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

func (e *ErrorParse) Error() string { return e.message }

func NewInvalidRequestError(message string) *ErrorInvalidRequest {
	return &ErrorInvalidRequest{
		message: message,
	}
}

// received message isn't a valid request
type ErrorInvalidRequest struct{ message string }

func (e *ErrorInvalidRequest) ErrorCode() int { return -32600 }

func (e *ErrorInvalidRequest) Error() string { return e.message }

// received message is invalid
type ErrorInvalidMessage struct{ message string }

func (e *ErrorInvalidMessage) ErrorCode() int { return -32700 }

func (e *ErrorInvalidMessage) Error() string { return e.message }

func NewInvalidParamsError(message string) *ErrorInvalidMessage {
	return &ErrorInvalidMessage{
		message: message,
	}
}

// unable to decode supplied params, or an invalid number of parameters
type ErrorInvalidParams struct{ message string }

func (e *ErrorInvalidParams) ErrorCode() int { return -32602 }

func (e *ErrorInvalidParams) Error() string { return e.message }
