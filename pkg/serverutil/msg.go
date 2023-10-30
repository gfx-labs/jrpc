package serverutil

import (
	"encoding/json"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

type Bundle struct {
	Messages []*jsonrpc.Message
	Batch    bool
}

func ParseBundle(raw json.RawMessage) *Bundle {
	a, b := jsonrpc.ParseMessage(raw)
	return &Bundle{
		Messages: a,
		Batch:    b,
	}
}
