package argreflect

import (
	"encoding/json"
	"errors"
	"fmt"
	"gfx.cafe/open/jrpc/contrib/codecs/websocket/wsjson"
	"reflect"
)

var jzon = wsjson.JZON

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
		return nil, errors.New("non-array args")
	}
	// Set any missing args to nil.
	for i := len(args); i < len(types); i++ {
		if types[i].Kind() != reflect.Ptr {
			return nil, fmt.Errorf("missing value for required argument %d", i)
		}
		args = append(args, reflect.Zero(types[i]))
	}
	return args, nil
}

func parseArgumentArray(p json.RawMessage, types []reflect.Type) ([]reflect.Value, error) {
	dec := jzon.BorrowIterator(p)
	defer jzon.ReturnIterator(dec)
	args := make([]reflect.Value, 0, len(types))
	for i := 0; dec.ReadArray(); i++ {
		if i >= len(types) {
			return args, fmt.Errorf("too many arguments, want at most %d", len(types))
		}
		argval := reflect.New(types[i])
		dec.ReadVal(argval.Interface())
		if err := dec.Error; err != nil {
			return args, fmt.Errorf("invalid argument %d: %v", i, err)
		}
		if argval.IsNil() && types[i].Kind() != reflect.Ptr {
			return args, fmt.Errorf("missing value for required argument %d", i)
		}
		args = append(args, argval.Elem())
	}
	return args, nil
}
