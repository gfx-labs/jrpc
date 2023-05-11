package jrpc

import (
	"context"

	"gfx.cafe/open/jrpc/codec"
)

type session struct {
}

func (s *session) MaxProcs() int {
	return 8
}

func (s *session) handle(ctx context.Context, stream codec.ReaderWriter, service Handler) error {
	msg, err := stream.ReadBatch(ctx)
	if err != nil {
		//TODO: deal with this error
		return err
	}
	messages, batch := codec.ParseMessage(msg)
	totalReplies := 0
	for _, v := range messages {
		if v.ID != nil && !v.ID.IsNull() {
			totalReplies = totalReplies + 1
		}
	}
	for _, msg := range messages {
		s.handleMessage(err)
	}
	return nil
}

func (s *session) handleMessage(ctx context.Context, stream codec.ReaderWriter, service Handler, msg *codec.Message) func() {
	return nil
}
