package serverutil

import (
	"encoding/json"

	"gfx.cafe/open/jrpc/pkg/codec"
)

type Bundle struct {
	Messages []*codec.Message
	Batch    bool
}

func ParseBundle(raw json.RawMessage) *Bundle {
	a, b := codec.ParseMessage(raw)
	return &Bundle{
		Messages: a,
		Batch:    b,
	}
}
