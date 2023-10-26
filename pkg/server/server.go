package server

import (
	"context"
	"sync"

	"gfx.cafe/open/jrpc/pkg/codec"

	"gfx.cafe/util/go/bufpool"

	"github.com/go-faster/jx"
	"github.com/goccy/go-json"
)

// Server is an RPC server.
// it is in charge of calling the handler on the message object, the json encoding of responses, and dealing with batch semantics.
// a server can be used to listenandserve multiple codecs at a time
type Server struct {
	services codec.Handler
}

// NewServer creates a new server instance with no registered handlers.
func NewServer(r codec.Handler) *Server {
	server := &Server{services: r}
	return server
}

// ServeCodec reads incoming requests from codec, calls the appropriate callback and writes
// the response back using the given codec. It will block until the codec is closed
func (s *Server) ServeCodec(ctx context.Context, remote codec.ReaderWriter) error {
	defer remote.Close()
	responder := &callResponder{
		remote: remote,
	}
	// add a cancel to the context so we can cancel all the child tasks on return
	ctx, cn := context.WithCancel(ContextWithPeerInfo(ctx, remote.PeerInfo()))
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
			go func() {
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
	err := c.remote.Send(func(e *jx.Encoder) error {
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
		err = codec.MarshalMessage(msg, e)
		if err != nil {
			return err
		}
		return nil
	})
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
			return c.remote.Send(func(e *jx.Encoder) error { return nil })
		}
	}
	// create the streaming encoder
	err = c.remote.Send(func(enc *jx.Encoder) error {
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
			m := msg
			enc.Obj(func(e *jx.Encoder) {
				e.Field("jsonrpc", func(e *jx.Encoder) {
					e.Str("2.0")
				})
				if m.ID != nil {
					e.Field("id", func(e *jx.Encoder) {
						e.Raw(m.ID.RawMessage())
					})
				}
				if m.Method != "" {
					e.Field("method", func(e *jx.Encoder) {
						e.Str(m.Method)
					})
				}
				for _, v := range m.ExtraFields {
					e.Field(v.Name, func(e *jx.Encoder) {
						e.Raw(v.Value)
					})
				}
				if m.Error != nil {
					e.Field("error", func(e *jx.Encoder) {
						codec.EncodeError(e, m.Error)
					})
				} else {
					// if there is no error, we try to marshal the result
					e.Field("result", func(e *jx.Encoder) {
						if v.dat != nil {
							err = json.NewEncoder(e).EncodeWithOption(v.dat, func(eo *json.EncodeOption) {
								eo.DisableNewline = true
							})
							if err != nil {
							}
						} else {
							e.Null()
						}
					})
				}
			})
		}
		if env.batch {
			enc.ArrEnd()
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
