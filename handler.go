// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package jrpc

import (
	"context"
	"encoding/json"
	"sync"

	"gfx.cafe/open/jrpc/codec"
)

type requestOp struct {
	ids  []json.RawMessage
	err  error
	resp chan json.RawMessage // receives up to len(ids) responses
}

type handler struct {
	reg        Handler
	respWait   map[string]*requestOp // active client requests
	callWG     sync.WaitGroup        // pending call goroutines
	rootCtx    context.Context       // canceled by close()
	cancelRoot func()                // cancel function for rootCtx
	conn       codec.Writer          // where responses will be sent

	peer codec.PeerInfo
}

type callProc struct {
	ctx context.Context
}

func newHandler(connCtx context.Context, conn codec.Writer, reg Handler) *handler {
	rootCtx, cancelRoot := context.WithCancel(connCtx)
	h := &handler{
		peer:       PeerInfoFromContext(connCtx),
		reg:        reg,
		conn:       conn,
		respWait:   make(map[string]*requestOp),
		rootCtx:    rootCtx,
		cancelRoot: cancelRoot,
	}
	return h
}

// handleBatch executes all messages in a batch and returns the responses.
func (h *handler) handleBatch(msgs []json.RawMessage) {
	// Emit error response for empty batches:
	if len(msgs) == 0 {
		h.startCallProc(func(cp *callProc) {
			h.conn.WriteJSON(cp.ctx, codec.ErrorMessage(codec.NewInvalidRequestError("empty batch")))
		})
		return
	}
	// Handle non-call messages first:
	calls := make([]json.RawMessage, 0, len(msgs))
	for _, msg := range msgs {
		//TODO: filter immediate handled
		//if handled := h.handleImmediate(msg); !handled {
		calls = append(calls, msg)
		//}
	}
	if len(calls) == 0 {
		return
	}
	// TODO: implement this
	// Process calls on a goroutine because they may block indefinitely:
	h.startCallProc(func(cp *callProc) {
		answers := make([]*json.RawMessage, 0, len(msgs))
		for _, msg := range calls {
			_ = msg
		}
		if len(answers) > 0 {
			h.conn.WriteJSON(cp.ctx, answers)
		}
	})
}

// handleMsg handles a single message.
func (h *handler) handleMsg(msg *codec.Message) {
	// TODO: implement this
	h.startCallProc(func(cp *callProc) {
		//	r := NewMsgRequest(cp.ctx, h.peer, *msg)
		//r := &Request{}
		//answer := h.handleCallMsg(cp, r)
		//if answer != nil {
		//	h.conn.WriteJSON(cp.ctx, answer)
		//}
	})
}

// close cancels all requests except for inflightReq and waits for
// call goroutines to shut down.
func (h *handler) close(err error, inflightReq *requestOp) {
	h.cancelAllRequests(err, inflightReq)
	h.callWG.Wait()
	h.cancelRoot()
}

// addRequestOp registers a request operation.
func (h *handler) addRequestOp(op *requestOp) {
	for _, id := range op.ids {
		h.respWait[string(id)] = op
	}
}

// removeRequestOps stops waiting for the given request IDs.
func (h *handler) removeRequestOp(op *requestOp) {
	for _, id := range op.ids {
		delete(h.respWait, string(id))
	}
}

// cancelAllRequests unblocks and removes pending requests and active subscriptions.
func (h *handler) cancelAllRequests(err error, inflightReq *requestOp) {
	didClose := make(map[*requestOp]bool)
	if inflightReq != nil {
		didClose[inflightReq] = true
	}

	for id, op := range h.respWait {
		// Remove the op so that later calls will not close op.resp again.
		delete(h.respWait, id)
		if !didClose[op] {
			op.err = err
			close(op.resp)
			didClose[op] = true
		}
	}
}

// startCallProc runs fn in a new goroutine and starts tracking it in the h.calls wait group.
func (h *handler) startCallProc(fn func(*callProc)) {
	h.callWG.Add(1)
	go func() {
		ctx, cancel := context.WithCancel(h.rootCtx)
		defer h.callWG.Done()
		defer cancel()
		fn(&callProc{ctx: ctx})
	}()
}

func (h *handler) handleCall(cp *callProc, r *Request) *Response {
	mw := NewReaderResponseWriterMsg(r)
	h.reg.ServeRPC(mw, r)
	return mw.Response()
}
