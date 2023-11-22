package clientutil

import (
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
