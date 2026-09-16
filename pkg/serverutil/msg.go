package serverutil

import (
	"encoding/json"

	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
)

type SimpleBundle struct {
	Msgs  []*jsonrpc.Message
	Batch bool
}

func ParseBundle(raw json.RawMessage) *SimpleBundle {
	a, b := jsonrpc.ParseMessage(raw)
	return &SimpleBundle{
		Msgs:  a,
		Batch: b,
	}
}

func (b *SimpleBundle) Messages() []*jsonrpc.Message {
	return b.Msgs
}

func (b *SimpleBundle) IsBatch() bool {
	return b.Batch
}
