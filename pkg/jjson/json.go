package jjson

import (
	"bytes"
	"encoding/json"
	"io"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/bytebufferpool"
)

var jConfig = jsoniter.Config{
	ValidateJsonRawMessage: false,
	EscapeHTML:             false,
	SortMapKeys:            true,
}.Froze()

func MarshalAndEncode(w io.Writer, v any) error {
	d := bytebufferpool.Get()
	defer bytebufferpool.Put(d)
	err := Encode(d, v)
	if err != nil {
		return err
	}
	_, err = d.WriteTo(w)
	return err
}

func Encode(w io.Writer, v any) error {
	s := jConfig.BorrowStream(w)
	defer jConfig.ReturnStream(s)
	switch cast := (v).(type) {
	case func(e *jsoniter.Stream):
		cast(s)
		return s.Flush()
	case json.Marshaler:
		s.WriteVal(v)
		return s.Flush()
	case io.Reader:
		_, err := io.Copy(w, cast)
		if err != nil {
			return err
		}
		return nil
	default:
		s.WriteVal(v)
		return s.Flush()
	}
}

func Decode(r io.Reader, v any) error {
	d := jConfig.NewDecoder(r)
	switch cast := (v).(type) {
	case json.Unmarshaler:
		return d.Decode(v)
	case io.Writer:
		_, err := io.Copy(cast, r)
		if err != nil {
			return err
		}
		return nil
	default:
		return d.Decode(v)
	}
}

func Unmarshal(xs []byte, v any) error {
	return Decode(bytes.NewBuffer(xs), v)
}

func Marshal(v any) ([]byte, error) {
	out := &bytes.Buffer{}
	err := Encode(out, v)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
