package jsonrpc

import (
	"encoding/json"
	"io"

	"github.com/gfx-labs/jrpc/pkg/jjson"
)

func EncodeObject(wr io.Writer, dat any) error {
	// if is nil, just write null
	if dat == nil {
		_, err := wr.Write(Null)
		if err != nil {
			return err
		}
		return nil
	}
	// if is not nil, do switch statement
	switch cast := (dat).(type) {
	case json.RawMessage:
		if len(cast) == 0 {
			_, err := wr.Write(Null)
			if err != nil {
				return err
			}
		} else {
			_, err := wr.Write(cast)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return jjson.Encode(wr, cast)
	}
}
