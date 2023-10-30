package clientutil

import (
	"encoding/json"
	"fmt"

	"gfx.cafe/util/go/generic"

	"gfx.cafe/open/jrpc/pkg/codec"
)

var msgPool = generic.HookPool[*codec.Message]{
	New: func() *codec.Message {
		return &codec.Message{}
	},
	FnPut: func(msg *codec.Message) {
		*msg = codec.Message{}
	},
}

func GetMessage() *codec.Message {
	return msgPool.Get()
}

func PutMessage(x *codec.Message) {
	msgPool.Put(x)
}

func FillBatch(ids map[int]int, msgs []*codec.Message, b []*codec.BatchElem) {
	answers := make(map[int]*codec.Message, len(msgs))
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
