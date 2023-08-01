package redis

import "errors"

const (
	// NOTE: if you change this, you will have to change the thing in jrpctest... its what its for now until tests get refactored
	maxRequestContentLength = 1024 * 1024 * 5
	contentType             = "application/json"
)

// https://www.jsonrpc.org/historical/json-rpc-over-http.html#id13
var acceptedContentTypes = []string{
	// https://www.jsonrpc.org/historical/json-rpc-over-http.html#id13
	contentType, "application/json-rpc", "application/jsonrequest",
	// these are added because they make sense, fight me!
	"application/jsonrpc2", "application/json-rpc2", "application/jrpc",
}
var ErrInvalidContentType = errors.New("invalid content type")
