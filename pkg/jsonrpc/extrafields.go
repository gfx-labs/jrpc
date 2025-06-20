package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// ExtraFields represents extra fields to be added to JSON-RPC messages.
// It functions similarly to http.Header.
type ExtraFields map[string]any

// Set sets the field associated with key to value.
// It replaces any existing values.
// Reserved JSON-RPC 2.0 fields ("jsonrpc", "id", "method", "params", "result", "error") cannot be set.
func (e ExtraFields) Set(key string, value any) {
	if isReservedField(key) {
		return
	}
	e[key] = value
}

// isReservedField checks if a field name is reserved by JSON-RPC 2.0
func isReservedField(key string) bool {
	switch key {
	case "jsonrpc", "id", "method", "params", "result", "error":
		return true
	default:
		return false
	}
}

// Get gets the value associated with the given key.
// If there are no values associated with the key, Get returns nil, false.
func (e ExtraFields) Get(key string) (any, bool) {
	v, ok := e[key]
	return v, ok
}

// Del deletes the values associated with key.
func (e ExtraFields) Del(key string) {
	delete(e, key)
}

// Clone returns a copy of e or nil if e is nil.
func (e ExtraFields) Clone() ExtraFields {
	if e == nil {
		return nil
	}
	e2 := make(ExtraFields, len(e))
	for k, v := range e {
		e2[k] = v
	}
	return e2
}

// toRawMessages converts the ExtraFields to a map of json.RawMessage
// for use in message marshaling.
func (e ExtraFields) toRawMessages() (map[string]json.RawMessage, error) {
	if len(e) == 0 {
		return nil, nil
	}
	result := make(map[string]json.RawMessage, len(e))
	for k, v := range e {
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal extra field %q: %w", k, err)
		}
		result[k] = json.RawMessage(data)
	}
	return result, nil
}