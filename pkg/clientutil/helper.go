package clientutil

import (
	"encoding/json"
	"fmt"
	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/util/go/generic"
)

var msgPool = generic.HookPool[*codec.Message]{
	New: func() *codec.Message {
		return &codec.Message{}
	},
}

func GetMessage() *codec.Message {
	return msgPool.Get()
}

func PutMessage(x *codec.Message) {
	msgPool.Put(x)
}

func FillBatch(ids []int, msgs []*codec.Message, b []*codec.BatchElem) {
	answers := map[int]*codec.Message{}
	for _, v := range msgs {
		answers[v.ID.Number()] = v
	}
	for i := range ids {
		idx := i
		ans, ok := answers[i]
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
		err := json.Unmarshal(ans.Result, b[idx].Result)
		if err != nil {
			b[idx].Error = err
		}
	}
}
