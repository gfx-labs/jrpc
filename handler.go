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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

	"tuxpa.in/a/zlog"
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
	conn       JsonWriter            // where responses will be sent
	log        *zlog.Logger

	subLock       sync.RWMutex
	clientSubs    map[string]*ClientSubscription // active client subscriptions
	serverSubs    map[SubID]*Subscription
	unsubscribeCb *callback

	peer PeerInfo
}

type callProc struct {
	ctx       context.Context
	notifiers []*Notifier
}

func newHandler(connCtx context.Context, conn JsonWriter, reg Router) *handler {
	rootCtx, cancelRoot := context.WithCancel(connCtx)
	h := &handler{
		peer:       PeerInfoFromContext(connCtx),
		reg:        reg,
		conn:       conn,
		respWait:   make(map[string]*requestOp),
		clientSubs: map[string]*ClientSubscription{},
		serverSubs: map[SubID]*Subscription{},
		rootCtx:    rootCtx,
		cancelRoot: cancelRoot,
		log:        zlog.Ctx(connCtx),
	}
	if h.peer.RemoteAddr != "" {
		cl := h.log.With().Str("conn", conn.RemoteAddr()).Logger()
		h.log = &cl
	}
	h.unsubscribeCb = newCallback(reflect.Value{}, reflect.ValueOf(h.unsubscribe))
	return h
}

// handleBatch executes all messages in a batch and returns the responses.
func (h *handler) handleBatch(msgs []*jsonrpcMessage) {
	// Emit error response for empty batches:
	if len(msgs) == 0 {
		h.startCallProc(func(cp *callProc) {
			h.conn.WriteJSON(cp.ctx, errorMessage(&invalidRequestError{"empty batch"}))
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
			r := NewMsgRequest(cp.ctx, h.peer, *msg)
			if answer := h.handleCallMsg(cp, r); answer != nil {
				answers = append(answers, answer)
			}
		}
		h.addSubscriptions(cp.notifiers)
		if len(answers) > 0 {
			h.conn.WriteJSON(cp.ctx, answers)
		}
		for _, n := range cp.notifiers {
			n.activate()
		}
	})
}

// handleMsg handles a single message.
func (h *handler) handleMsg(msg *jsonrpcMessage) {
	if ok := h.handleImmediate(msg); ok {
		return
	}
	h.startCallProc(func(cp *callProc) {
		r := NewMsgRequest(cp.ctx, h.peer, *msg)
		answer := h.handleCallMsg(cp, r)
		h.addSubscriptions(cp.notifiers)
		if answer != nil {
			h.conn.WriteJSON(cp.ctx, answer)
		}
		for _, n := range cp.notifiers {
			n.activate()
		}
	})
}

// close cancels all requests except for inflightReq and waits for
// call goroutines to shut down.
func (h *handler) close(err error, inflightReq *requestOp) {
	h.cancelAllRequests(err, inflightReq)
	h.callWG.Wait()
	h.cancelRoot()
	h.cancelServerSubscriptions(err)
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
	start := time.Now()
	switch {
	case msg.isNotification():
		if strings.HasSuffix(msg.Method, notificationMethodSuffix) {
			h.handleSubscriptionResult(msg)
			return true
		}
		return false
	case msg.isResponse():
		h.handleResponse(msg)
		h.log.Trace().Str("reqid", string(msg.ID.RawMessage())).Dur("duration", time.Since(start)).Msg("Handled RPC response")
		return true
	default:
		return false
	}
}

func (h *handler) handleSubscriptionResult(msg *jsonrpcMessage) {
	var result subscriptionResult
	if err := json.Unmarshal(msg.Params, &result); err != nil {
		h.log.Trace().Msg("Dropping invalid subscription message")
		return
	}
	if h.clientSubs[result.ID] != nil {
		h.clientSubs[result.ID].deliver(result.Result)
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
	if op.sub == nil {
		// not a sub, so just send the msg back
		op.resp <- msg
		return
	}
	// For subscription responses, start the subscription if the server
	// indicates success. EthSubscribe gets unblocked in either case through
	// the op.resp channel.
	defer close(op.resp)
	if msg.Error != nil {
		op.err = msg.Error
		return
	}
	if op.err = json.Unmarshal(msg.Result, &op.sub.subid); op.err == nil {
		go op.sub.start()
		h.clientSubs[op.sub.subid] = op.sub
	}
}

// handleCallMsg executes a call message and returns the answer.
// TODO: export prometheus metrics maybe? also fix logging
func (h *handler) handleCallMsg(ctx *callProc, r *Request) *jsonrpcMessage {
	switch {
	case r.isNotification():
		go h.handleCall(ctx, r)
		return nil
	case r.isCall():
		resp := h.handleCall(ctx, r)
		return resp
	case r.hasValidID():
		return r.makeError(&invalidRequestError{"invalid request"})
	default:
		return errorMessage(&invalidRequestError{"invalid request"})
	}
}

func (h *handler) handleCall(cp *callProc, r *Request) *jsonrpcMessage {
	callb := h.reg.Match(NewRouteContext(), r.Method)
	mw := NewReaderResponseWriterMsg(r)
	if r.isSubscribe() {
		return h.handleSubscribe(cp, r)
	}
	if r.isUnsubscribe() {
		h.unsubscribeCb.ServeRPC(mw, r)
		return mw.Msg()
	}
	// no method found
	msg := r.Msg()
	if !callb {
		return msg.errorResponse(&methodNotFoundError{method: r.Method})
	}
	// now actually run the handler
	h.reg.ServeRPC(mw, r)
	return mw.Msg()
}

// handleSubscribe processes *_subscribe method calls.
func (h *handler) handleSubscribe(cp *callProc, r *Request) *jsonrpcMessage {
	switch h.peer.Transport {
	case "http", "https":
		return r.makeError(ErrNotificationsUnsupported)
	}

	// Subscription method name is first argument.
	name, err := parseSubscriptionName(r.Params)
	if err != nil {
		return r.makeError(&invalidParamsError{err.Error()})
	}
	namespace := r.namespace()
	has := h.reg.Match(NewRouteContext(), r.Method)
	if !has {
		return r.makeError(&subscriptionNotFoundError{namespace, name})
	}
	// Install notifier in context so the subscription handler can find it.
	n := &Notifier{h: h, namespace: namespace, idgen: randomIDGenerator()}
	cp.notifiers = append(cp.notifiers, n)
	req := r.WithContext(cp.ctx)
	// now actually run the handler
	req = req.WithContext(
		context.WithValue(req.ctx, notifierKey{}, n),
	)
	mw := NewReaderResponseWriterMsg(req)
	h.reg.ServeRPC(mw, req)
	return mw.Msg()
}

// parseSubscriptionName extracts the subscription name from an encoded argument array.
func parseSubscriptionName(rawArgs json.RawMessage) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(rawArgs))
	if tok, _ := dec.Token(); tok != json.Delim('[') {
		return "", errors.New("non-array args")
	}
	v, _ := dec.Token()
	method, ok := v.(string)
	if !ok {
		return "", errors.New("expected subscription name as first argument")
	}
	return method, nil
}

func (h *handler) unsubscribe(ctx context.Context, id SubID) (bool, error) {
	h.subLock.Lock()
	defer h.subLock.Unlock()

	s := h.serverSubs[id]
	if s == nil {
		return false, ErrSubscriptionNotFound
	}
	close(s.err)
	delete(h.serverSubs, id)
	return true, nil
}

// cancelServerSubscriptions removes all subscriptions and closes their error channels.
func (h *handler) cancelServerSubscriptions(err error) {
	h.subLock.Lock()
	defer h.subLock.Unlock()

	for id, s := range h.serverSubs {
		s.err <- err
		close(s.err)
		delete(h.serverSubs, id)
	}
}

func (h *handler) addSubscriptions(nn []*Notifier) {
	h.subLock.Lock()
	defer h.subLock.Unlock()

	for _, n := range nn {
		if sub := n.takeSubscription(); sub != nil {
			h.serverSubs[sub.ID] = sub
		}
	}
}
