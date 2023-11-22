package jjson

import (
	"bytes"
	"io"

	jsoniter "github.com/json-iterator/go"
)

var encPool = NewPool()

var jConfig = jsoniter.Config{
	ValidateJsonRawMessage: false,
	EscapeHTML:             false,
	SortMapKeys:            true,
}.Froze()

func MarshalAndEncode(w io.Writer, v any) error {
	d := encPool.Get()
	defer encPool.Put(d)
	err := Encode(w, v)
	if err != nil {
		return err
	}
	_, err = d.WriteTo(w)
	return err
}

func Encode(w io.Writer, v any) error {
	s := jConfig.BorrowStream(w)
	defer jConfig.ReturnStream(s)
	s.WriteVal(v)
	return s.Flush()
}

func Decode(r io.Reader, v any) error {
	d := jConfig.NewDecoder(r)
	return d.Decode(v)
}

func Unmarshal(xs []byte, v any) error {
	d := jConfig.NewDecoder(bytes.NewBuffer(xs))
	return d.Decode(v)
}

func Marshal(v any) ([]byte, error) {
	out := &bytes.Buffer{}
	s := jConfig.BorrowStream(out)
	defer jConfig.ReturnStream(s)
	s.WriteVal(v)
	err := s.Flush()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
