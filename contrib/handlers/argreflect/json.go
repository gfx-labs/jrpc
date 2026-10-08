package argreflect

import (
	"encoding/json"
	"encoding/json/jsontext"
	"fmt"
	"reflect"

	"github.com/gfx-labs/jrpc/pkg/jjson"
	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
)

// parsePositionalArguments tries to parse the given args to an array of values with the
// given types. It returns the parsed values or an error when the args could not be
// parsed. Missing optional arguments are returned as reflect.Zero values.
func parsePositionalArguments(rawArgs json.RawMessage, types []reflect.Type) ([]reflect.Value, error) {
	var args []reflect.Value
	switch {
	case len(rawArgs) == 0:
	case rawArgs[0] == '[':
		// Read argument array.
		var err error
		if args, err = parseArgumentArray(rawArgs, types); err != nil {
			return nil, err
		}
	case string(rawArgs) == "null":
		return nil, nil
	default:
		return nil, jsonrpc.NewInvalidParamsError("non-array args")
	}
	// Set any missing args to nil.
	for i := len(args); i < len(types); i++ {
		if types[i].Kind() != reflect.Ptr {
			return nil, jsonrpc.NewInvalidParamsError(fmt.Sprintf("missing value for required argument %d", i))
		}
		args = append(args, reflect.Zero(types[i]))
	}
	return args, nil
}

func parseArgumentArray(p json.RawMessage, types []reflect.Type) ([]reflect.Value, error) {
	d := jjson.GetDecoder(p)
	defer jjson.PutDecoder(d)
	dec := &d.Decoder
	args := make([]reflect.Value, 0, len(types))
	if tok, err := dec.ReadToken(); err != nil || tok.Kind() != jsontext.KindBeginArray {
		return args, jsonrpc.NewInvalidParamsError("expected array")
	}
	i := 0
	for dec.PeekKind() != jsontext.KindEndArray {
		if i >= len(types) {
			return args, jsonrpc.NewInvalidParamsError(fmt.Sprintf("too many arguments, want at most %d", len(types)))
		}
		argval := reflect.New(types[i])
		raw, err := dec.ReadValue()
		if err != nil {
			return args, jsonrpc.NewInvalidParamsError(fmt.Sprintf("invalid raw argument %d: %v", i, err))
		}
		err = jjson.Unmarshal(raw, argval.Interface())
		if err != nil {
			return args, jsonrpc.NewInvalidParamsError(fmt.Sprintf("invalid argument %d: %v", i, err))
		}
		if argval.IsNil() && types[i].Kind() != reflect.Ptr {
			return nil, jsonrpc.NewInvalidParamsError(fmt.Sprintf("missing value for required argument %d", i))
		}
		args = append(args, argval.Elem())
		i++
	}
	return args, nil
}
