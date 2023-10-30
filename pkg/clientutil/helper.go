package clientutil

import (
	"encoding/json"
	"fmt"

	"gfx.cafe/util/go/generic"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

var msgPool = generic.HookPool[*jsonrpc.Message]{
	New: func() *jsonrpc.Message {
		return &jsonrpc.Message{}
	},
	FnPut: func(msg *jsonrpc.Message) {
		*msg = jsonrpc.Message{}
	},
}

func GetMessage() *jsonrpc.Message {
	return msgPool.Get()
}

func PutMessage(x *jsonrpc.Message) {
	msgPool.Put(x)
}

func FillBatch(ids map[int]int, msgs []*jsonrpc.Message, b []*jsonrpc.BatchElem) {
	answers := make(map[int]*jsonrpc.Message, len(msgs))
	for _, v := range msgs {
		answers[v.ID.Number()] = v
	}
	for idx, id := range ids {
		ans, ok := answers[id]
		if !ok {
			b[idx].Error = fmt.Errorf("No response found")
			continue
		}
		if ans.Error != nil {
			b[idx].Error = ans.Error
			continue
		}
		if b[idx].Result == nil {
			continue
		}
		err := json.NewDecoder(ans.Result).Decode(b[idx].Result)
		if err != nil {
			b[idx].Error = err
		}
	}
}
