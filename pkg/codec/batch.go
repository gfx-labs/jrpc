package codec

// BatchElem is an element in a batch request.
type BatchElem struct {
	Method string
	Params any

	IsNotification bool

	// The result is unmarshaled into this field. Result must be set to a
	// non-nil pointer value of the desired type, otherwise the response will be
	// discarded.
	Result any
	// Error is set if the server returns an error for this request, or if
	// unmarshaling into Result fails. It is not set for I/O errors.
	Error error
}
