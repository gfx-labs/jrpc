package server

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/util/mapset"

	"gfx.cafe/util/go/bufpool"

	"github.com/go-faster/jx"
	"github.com/goccy/go-json"
)

// Server is an RPC server.
// it is in charge of calling the handler on the message object, the json encoding of responses, and dealing with batch semantics.
// a server can be used to listenandserve multiple codecs at a time
type Server struct {
	services codec.Handler
	run      int32
	codecs   *mapset.Set[codec.ReaderWriter]
	Tracing  Tracing
}

type Tracing struct {
}

// NewServer creates a new server instance with no registered handlers.
func NewServer(r codec.Handler) *Server {
	server := &Server{
		codecs: mapset.NewSet[codec.ReaderWriter](),
		run:    1,
	}
	server.services = r
	return server
}

func (s *Server) serveBatch(ctx context.Context,
	incoming []*codec.Message,
	batch bool,
	remote codec.ReaderWriter, responder *callResponder) error {
	env := &callEnv{
		batch: batch,
	}

	// check for empty batch
	if batch && len(incoming) == 0 {
		// if it is empty batch, send the empty batch error and immediately return
		return responder.send(ctx, &callEnv{
			responses: []*callRespWriter{{
				pkt: &codec.Message{
					ID:    codec.NewNullIDPtr(),
					Error: codec.NewInvalidRequestError("empty batch"),
				},
			}},
			batch: false,
		})
	}

	// populate the envelope we are about to send. this is synchronous pre-prpcessing
	for _, v := range incoming {
		// create the response writer
		rw := &callRespWriter{
			notifications: func(env *notifyEnv) error { return responder.notify(ctx, env) },
			header:        remote.PeerInfo().HTTP.Headers,
		}
		env.responses = append(env.responses, rw)
		// a nil incoming message means an empty response
		if v == nil {
			rw.msg = &codec.Message{ID: codec.NewNullIDPtr()}
			rw.pkt = &codec.Message{ID: codec.NewNullIDPtr()}
			continue
		}
		rw.msg = v
		if v.ID == nil {
			rw.pkt = &codec.Message{ID: codec.NewNullIDPtr()}
			continue
		}
		rw.pkt = &codec.Message{ID: v.ID}
	}

	// create a waitgroup
	wg := sync.WaitGroup{}
	wg.Add(len(env.responses))
	// for each item in the envelope
	peerInfo := remote.PeerInfo()
	for _, vRef := range env.responses {
		v := vRef
		// process each request in its own goroutine
		go func() {
			defer wg.Done()
			// early respond to nil requests
			if v.msg == nil || len(v.msg.Method) == 0 {
				v.pkt.Error = codec.NewInvalidRequestError("invalid request")
				return
			}
			if v.msg.ID == nil || v.msg.ID.IsNull() {
				// it's a notification, so we mark skip and we don't write anything for it
				v.skip = true
				return
			}
			r := codec.NewRequestFromMessage(
				ctx,
				v.msg,
			)
			r.Peer = peerInfo
			s.services.ServeRPC(v, r)
		}()
	}
	wg.Wait()
	return responder.send(ctx, env)
}

// ServeCodec reads incoming requests from codec, calls the appropriate callback and writes
// the response back using the given codec. It will block until the codec is closed or the
// server is stopped. In either case the codec is closed when this function returns.
func (s *Server) ServeCodec(pctx context.Context, remote codec.ReaderWriter) error {
	defer remote.Close()
	// Don't serve if server is stopped.
	if atomic.LoadInt32(&s.run) == 0 {
		return fmt.Errorf("Server stopped")
	}
	// Add the codec to the set so it can be closed by Stop.
	s.codecs.Add(remote)
	defer s.codecs.Remove(remote)
	responder := &callResponder{
		remote: remote,
	}
	// add a cancel to the context so we can cancel all the child tasks on return
	ctx, cn := context.WithCancel(ContextWithPeerInfo(pctx, remote.PeerInfo()))
	defer cn()

	errch := make(chan error)
	go func() {
		for {
			// read messages from the stream synchronously
			incoming, batch, err := remote.ReadBatch(ctx)
			if err != nil {
				errch <- err
				return
			}
			// process each in a goroutine
			go func() {
				// the only reason this should error is if
				err = s.serveBatch(ctx, incoming, batch, remote, responder)
				if err != nil {
					errch <- err
					return
				}
			}()
		}
	}()
	// exit on either the first error, or the context closing.
	select {
	case <-ctx.Done():
		return nil
	case err := <-errch:
		return err
	}
}

// Stop stops reading new requests, waits for stopPendingRequestTimeout to allow pending
// requests to finish, then closes all codecs which will cancel pending requests and
// subscriptions.
func (s *Server) Stop() {
	if atomic.CompareAndSwapInt32(&s.run, 1, 0) {
		s.codecs.Each(func(c codec.ReaderWriter) bool {
			c.Close()
			return true
		})
	}
}

type callResponder struct {
	remote codec.ReaderWriter
	mu     sync.Mutex
}

type notifyEnv struct {
	method string
	dat    any
	extra  []codec.RequestField
}

func (c *callResponder) notify(ctx context.Context, env *notifyEnv) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	enc := jx.GetEncoder()
	enc.Reset()
	enc.Grow(4096)
	defer jx.PutEncoder(enc)
	//enc := jx.NewStreamingEncoder(c.remote, 4096)
	msg := &codec.Message{}
	var err error
	//  allocate a temp buffer for this packet
	buf := bufpool.GetStd()
	defer bufpool.PutStd(buf)
	err = json.NewEncoder(buf).Encode(env.dat)
	if err != nil {
		msg.Error = err
	} else {
		msg.Params = buf.Bytes()
	}
	msg.ExtraFields = env.extra
	// add the method
	msg.Method = env.method
	err = codec.MarshalMessage(msg, enc)
	if err != nil {
		return err
	}
	err = c.remote.Send(ctx, enc.Bytes())
	if err != nil {
		return err
	}
	return nil
}

type callEnv struct {
	responses []*callRespWriter
	batch     bool
}

func (c *callResponder) send(ctx context.Context, env *callEnv) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// notification gets nothing
	// if all msgs in batch are notification, we trigger an allSkip and write nothing
	if env.batch {
		allSkip := true
		for _, v := range env.responses {
			if v.skip != true {
				allSkip = false
			}
		}
		if allSkip {
			return nil
		}
	}
	// create the streaming encoder
	enc := jx.GetEncoder()
	enc.Reset()
	enc.Grow(4096)
	defer jx.PutEncoder(enc)
	if env.batch {
		enc.ArrStart()
	}
	for _, v := range env.responses {
		msg := v.pkt
		// if we are a batch AND we are supposed to skip, then continue
		// this means that for a non-batch notification, we do not skip! this is to ensure we get always a "response" for http-like endpoints
		if env.batch && v.skip {
			continue
		}
		// if there is no error, we try to marshal the result
		if msg.Error == nil {
			buf := bufpool.GetStd()
			defer bufpool.PutStd(buf)
			je := json.NewEncoder(buf)
			err = je.EncodeWithOption(v.dat)
			if err != nil {
				msg.Error = err
			} else {
				msg.Result = buf.Bytes()
				msg.Result = bytes.TrimSuffix(msg.Result, []byte{'\n'})
			}
		}
		// then marshal the whole message into the stream
		err := codec.MarshalMessage(msg, enc)
		if err != nil {
			return err
		}
	}
	if env.batch {
		enc.ArrEnd()
	}
	err = c.remote.Send(ctx, enc.Bytes())
	if err != nil {
		return err
	}
	return nil
}
