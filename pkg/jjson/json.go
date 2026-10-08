package jjson

import (
	"bytes"
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"io"
	"sync"

	"github.com/valyala/bytebufferpool"
)

// Options are the json/v2 options used to encode and decode user values (the v2 defaults).
var Options = json.JoinOptions()

// decodeOptions are used for raw message scanning. Duplicate names and invalid UTF-8 are passed through unchecked.
var decodeOptions = json.JoinOptions(
	jsontext.AllowDuplicateNames(true),
	jsontext.AllowInvalidUTF8(true),
)

// NewDecoder returns a streaming jsontext decoder using the package decode options.
func NewDecoder(r io.Reader) *jsontext.Decoder {
	return jsontext.NewDecoder(r, decodeOptions)
}

// Decoder is a pooled jsontext decoder reading directly from a byte slice.
type Decoder struct {
	jsontext.Decoder
	buf bytes.Buffer
}

var decoderPool = sync.Pool{New: func() any { return new(Decoder) }}

// GetDecoder returns a pooled decoder over data. The decoder does not copy data.
func GetDecoder(data []byte) *Decoder {
	d := decoderPool.Get().(*Decoder)
	d.buf = *bytes.NewBuffer(data)
	d.Reset(&d.buf, decodeOptions)
	return d
}

// PutDecoder returns a decoder to the pool.
func PutDecoder(d *Decoder) {
	d.buf = bytes.Buffer{}
	d.Reset(&d.buf)
	decoderPool.Put(d)
}

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

// Encode writes v to w. A func(*jsontext.Encoder) error is called with an encoder over w,
// and an io.Reader is copied as-is.
func Encode(w io.Writer, v any) error {
	switch cast := v.(type) {
	case func(e *jsontext.Encoder) error:
		enc := jsontext.NewEncoder(w, Options)
		return cast(enc)
	case io.Reader:
		_, err := io.Copy(w, cast)
		return err
	default:
		return json.MarshalWrite(w, v, Options)
	}
}

// Decode reads a single JSON value from r until EOF into v. An io.Writer target receives the raw bytes.
func Decode(r io.Reader, v any) error {
	if cast, ok := v.(io.Writer); ok {
		_, err := io.Copy(cast, r)
		return err
	}
	return json.UnmarshalRead(r, v, Options)
}

func Unmarshal(xs []byte, v any) error {
	if cast, ok := v.(io.Writer); ok {
		_, err := cast.Write(xs)
		return err
	}
	return json.Unmarshal(xs, v, Options)
}

func Marshal(v any) ([]byte, error) {
	switch v.(type) {
	case func(e *jsontext.Encoder) error, io.Reader:
		out := &bytes.Buffer{}
		if err := Encode(out, v); err != nil {
			return nil, err
		}
		return out.Bytes(), nil
	default:
		return json.Marshal(v, Options)
	}
}
