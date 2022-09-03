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
	"strconv"
	"sync"
	"time"

	"git.tuxpa.in/a/zlog"
)

// handler handles JSON-RPC messages. There is one handler per connection. Note that
// handler is not safe for concurrent use. Message handling never blocks indefinitely
// because RPCs are processed on background goroutines launched by handler.
//
// The entry points for incoming messages are:
//
//	h.handleMsg(message)
//	h.handleBatch(message)
//
// Outgoing calls use the requestOp struct. Register the request before sending it
// on the connection:
//
//	op := &requestOp{ids: ...}
//	h.addRequestOp(op)
//
// Now send the request, then wait for the reply to be delivered through handleMsg:
//
//	if err := op.wait(...); err != nil {
//	    h.removeRequestOp(op) // timeout, etc.
//	}
type handler struct {
	reg        Router
	respWait   map[string]*requestOp // active client requests
	callWG     sync.WaitGroup        // pending call goroutines
	rootCtx    context.Context       // canceled by close()
	cancelRoot func()                // cancel function for rootCtx
	conn       jsonWriter            // where responses will be sent
	log        *zlog.Logger

	peer PeerInfo
}

type callProc struct {
	ctx context.Context
}

func newHandler(connCtx context.Context, conn jsonWriter, reg Router) *handler {
	rootCtx, cancelRoot := context.WithCancel(connCtx)
	h := &handler{
		peer:       PeerInfoFromContext(connCtx),
		reg:        reg,
		conn:       conn,
		respWait:   make(map[string]*requestOp),
		rootCtx:    rootCtx,
		cancelRoot: cancelRoot,
		log:        zlog.Ctx(connCtx),
	}
	if h.peer.RemoteAddr != "" {
		cl := h.log.With().Str("conn", conn.remoteAddr()).Logger()
		h.log = &cl
	}
	if h.peer.RemoteAddr == "" {
		h.log.Error().Msg("CONNECTION WITHOUT REMOTE IP DETECTED. PLEASE MAKE SURE YOU KNOW WHAT YOU ARE DOING, OTHERWISE, THIS COULD BE A SECURITY ISSUE")
	}
	return h
}

// handleBatch executes all messages in a batch and returns the responses.
func (h *handler) handleBatch(msgs []*jsonrpcMessage) {
	// Emit error response for empty batches:
	if len(msgs) == 0 {
		h.startCallProc(func(cp *callProc) {
			h.conn.writeJSON(cp.ctx, errorMessage(&invalidRequestError{"empty batch"}))
		})
		return
	}

	// Handle non-call messages first:
	calls := make([]*jsonrpcMessage, 0, len(msgs))
	for _, msg := range msgs {
		if handled := h.handleImmediate(msg); !handled {
			calls = append(calls, msg)
		}
	}
	if len(calls) == 0 {
		return
	}
	// Process calls on a goroutine because they may block indefinitely:
	h.startCallProc(func(cp *callProc) {
		answers := make([]*jsonrpcMessage, 0, len(msgs))
		for _, msg := range calls {
			if answer := h.handleCallMsg(cp, msg); answer != nil {
				answers = append(answers, answer)
			}
		}
		if len(answers) > 0 {
			h.conn.writeJSON(cp.ctx, answers)
		}
	})
}

// handleMsg handles a single message.
func (h *handler) handleMsg(msg *jsonrpcMessage) {
	if ok := h.handleImmediate(msg); ok {
		return
	}
	h.startCallProc(func(cp *callProc) {
		answer := h.handleCallMsg(cp, msg)
		if answer != nil {
			h.conn.writeJSON(cp.ctx, answer)
		}
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

// handleImmediate executes non-call messages. It returns false if the message is a
// call or requires a reply.
func (h *handler) handleImmediate(msg *jsonrpcMessage) bool {
	start := NewTimer()
	switch {
	case msg.isNotification():
		return true
	case msg.isResponse():
		h.handleResponse(msg)
		h.log.Trace().Str("reqid", string(msg.ID.RawMessage())).Dur("duration", start.Since(start)).Msg("Handled RPC response")
		return true
	default:
		return false
	}
}

// handleResponse processes method call responses.
func (h *handler) handleResponse(msg *jsonrpcMessage) {
	op := h.respWait[string(msg.ID.RawMessage())]
	if op == nil {
		h.log.Debug().Str("reqid", string(msg.ID.RawMessage())).Msg("Unsolicited RPC response")
		return
	}
	delete(h.respWait, string(msg.ID.RawMessage()))
	op.resp <- msg
}

// handleCallMsg executes a call message and returns the answer.
// TODO: export prometheus metrics maybe? also fix logging
func (h *handler) handleCallMsg(ctx *callProc, msg *jsonrpcMessage) *jsonrpcMessage {
	// start := NewTimer()
	switch {
	case msg.isNotification():
		h.handleCall(ctx, msg)
		//	h.log.Debug().Str("method", msg.Method).Dur("duration", start.Since()).Msg("Served")
		return nil
	case msg.isCall():
		resp := h.handleCall(ctx, msg)
		// var ctx []any
		//		log2 := h.log.With()
		//		log2.Str("reqid", string(msg.ID)).Dur("duration", start.Since())
		if resp.Error != nil {
			//			log2.Str("err", resp.Error.Message)
			//			if resp.Error.Data != nil {
			//				log2.Interface("errdata", resp.Error.Data)
			//			}
			//		sl := log2.Logger()
			//		sl.Warn().Str("method", msg.Method).Interface("ctx", ctx).Msg("Served")
		} else {
			//			sl := log2.Logger()
			//			sl.Debug().Str("method", msg.Method).Interface("ctx", ctx).Msg("Served")
		}
		return resp
	case msg.hasValidID():
		return msg.errorResponse(&invalidRequestError{"invalid request"})
	default:
		return errorMessage(&invalidRequestError{"invalid request"})
	}
}

func (h *handler) handleCall(cp *callProc, msg *jsonrpcMessage) *jsonrpcMessage {
	callb := h.reg.Match(NewRouteContext(), msg.Method)
	// no method found
	if !callb {
		return msg.errorResponse(&methodNotFoundError{method: msg.Method})
	}

	start := time.Now()
	// TODO:  there's a copy here
	// TODO: add buffer pool here
	req := &Request{ctx: cp.ctx, msg: *msg, peer: h.peer}
	mw := NewReaderResponseWriterMsg(req)
	h.reg.ServeRPC(mw, req)
	{
		// Collect the statistics for RPC calls if metrics is enabled.
		// We only care about pure rpc call. Filter out subscription.
		rpcRequestGauge.Inc(1)
		if mw.msg.Error != nil {
			failedRequestGauge.Inc(1)
		} else {
			successfulRequestGauge.Inc(1)
		}
		rpcServingTimer.UpdateSince(start)
		updateServeTimeHistogram(msg.Method, mw.msg.Error == nil, time.Since(start))
	}
	return mw.msg
}

type idForLog struct{ json.RawMessage }

func (id idForLog) String() string {
	if s, err := strconv.Unquote(string(id.RawMessage)); err == nil {
		return s
	}
	return string(id.RawMessage)
}
