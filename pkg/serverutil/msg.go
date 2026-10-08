package serverutil

import (
	"encoding/json"

	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
)

type SimpleBundle struct {
	Msgs  []*jsonrpc.Message
	Batch bool
}

// ParseBundle parses raw as a (batch of) JSON-RPC message(s). Malformed JSON yields a bundle
// with a single parse error message, which the server answers with a -32700 response.
func ParseBundle(raw json.RawMessage) *SimpleBundle {
	a, b, err := jsonrpc.ParseMessage(raw)
	if err != nil {
		return ParseErrorBundle(err)
	}
	return &SimpleBundle{
		Msgs:  a,
		Batch: b,
	}
}

// ParseErrorBundle returns a bundle that makes the server respond with a parse error.
func ParseErrorBundle(err error) *SimpleBundle {
	return &SimpleBundle{
		Msgs: []*jsonrpc.Message{{
			ID:    jsonrpc.NewNullIDPtr(),
			Error: jsonrpc.NewParseError(err.Error()),
		}},
	}
}

func (b *SimpleBundle) Messages() []*jsonrpc.Message {
	return b.Msgs
}

func (b *SimpleBundle) IsBatch() bool {
	return b.Batch
}
