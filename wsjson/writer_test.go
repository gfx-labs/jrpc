package wsjson

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"testing"
)

type noTrailingNewlineWriter struct {
	w io.Writer
}

func (n *noTrailingNewlineWriter) Write(xs []byte) (int, error) {
	if xs[len(xs)-1] == '\n' {
		xs = xs[:len(xs)-1]
	}
	return n.w.Write(xs)
}

func TestNoTrailingNewlineWriter(t *testing.T) {
	buf := new(bytes.Buffer)
	wr := &noTrailingNewlineWriter{w: buf}
	enc := json.NewEncoder(wr)
	enc.Encode(map[string]any{"hi": "there", "how": "are", "you": "?"})
	enc.Encode(map[string]any{"hi": "there", "how": "are", "you": "?"})
	enc.Encode(map[string]any{"hi": "there", "how": "are", "you": "?"})
	log.Println(string(buf.Bytes()))
}
