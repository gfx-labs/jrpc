package server

import (
	"context"
	"errors"
	"sync"

	"gfx.cafe/open/jrpc/pkg/codec"
	"golang.org/x/sync/semaphore"

	"gfx.cafe/util/go/bufpool"

	"github.com/go-faster/jx"
	"github.com/goccy/go-json"
)

// Server is an RPC server.
// it is in charge of calling the handler on the message object, the json encoding of responses, and dealing with batch semantics.
// a server can be used to listenandserve multiple codecs at a time
type Server struct {
	services codec.Handler

	lctx context.Context
	cn   context.CancelFunc
}

// NewServer creates a new server instance with no registered handlers.
func NewServer(r codec.Handler) *Server {
	server := &Server{services: r}
	server.lctx, server.cn = context.WithCancel(context.Background())
	return server
}

// ServeCodec reads incoming requests from codec, calls the appropriate callback and writes
// the response back using the given codec. It will block until the codec is closed
func (s *Server) ServeCodec(ctx context.Context, remote codec.ReaderWriter) error {
	defer remote.Close()

	sema := semaphore.NewWeighted(1)
	// add a cancel to the context so we can cancel all the child tasks on return
	ctx, cn := context.WithCancel(ContextWithPeerInfo(ctx, remote.PeerInfo()))
	defer cn()

	allErrs := []error{}
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	err := func() error {
		for {
			// read messages from the stream synchronously
			incoming, batch, err := remote.ReadBatch(ctx)
			if err != nil {
				return err
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				responder := &callResponder{
					remote: remote,
					batch:  batch,
					mu:     sema,
				}
				err = s.serveBatch(ctx, incoming, responder)
				if err != nil {
					mu.Lock()
					defer mu.Unlock()
					allErrs = append(allErrs, err)
				}
			}()
		}
	}()
	wg.Wait()
	allErrs = append(allErrs, err)
	if len(allErrs) > 0 {
		return errors.Join(allErrs...)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) {
	s.cn()
}

func (s *Server) serveBatch(ctx context.Context,
	incoming []*codec.Message,
	r *callResponder,
) error {
	// check for empty batch
	if r.batch && len(incoming) == 0 {
		// if it is empty batch, send the empty batch error and immediately return
		err := r.mu.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		defer r.mu.Release(1)
		err = r.send(ctx, &callEnv{
			id:  codec.NewNullIDPtr(),
			err: codec.NewInvalidRequestError("empty batch"),
		})
		if err != nil {
			return err
		}
		err = r.remote.Flush()
		if err != nil {
			return err
		}
		return nil
	}

	rs := []*callRespWriter{}

	totalRequests := 0
	// populate the envelope we are about to send. this is synchronous pre-prpcessing
	for _, v := range incoming {
		// create the response writer
		rw := &callRespWriter{
			ctx: ctx,
			cr:  r,
		}
		rs = append(rs, rw)
		// a nil incoming message means return an invalid request.
		if v == nil {
			v = &codec.Message{ID: codec.NewNullIDPtr()}
			rw.err = codec.NewInvalidRequestError("invalid request")
		}
		rw.msg = v
		rw.msg.ExtraFields = codec.ExtraFields{}
		rw.msg.Error = nil
		// zero length method is always invalid request
		if len(v.Method) == 0 {
			// assume if the method is not there AND the id is not there that it's an invalid REQUEST not notification
			// this makes sure we add 1 to totalRequests
			if v.ID == nil {
				v.ID = codec.NewNullIDPtr()
			}
			rw.err = codec.NewInvalidRequestError("invalid request")
		}
		// requests and malformed requests both count as requests
		if v.ID != nil {
			totalRequests += 1
		}
	}
	var doneMu *semaphore.Weighted
	doneMu = semaphore.NewWeighted(int64(totalRequests))
	err := doneMu.Acquire(ctx, int64(totalRequests))
	if err != nil {
		return err
	}

	// create a waitgroup for everything
	wg := sync.WaitGroup{}
	wg.Add(len(rs))
	// for each item in the envelope
	peerInfo := r.remote.PeerInfo()
	batchResults := []*callRespWriter{}
	for _, vRef := range rs {
		v := vRef
		if r.batch {
			v.noStream = true
			if v.msg.ID != nil {
				v.doneMu = doneMu
				batchResults = append(batchResults, v)
			}
		}
		// now process each request in its own goroutine
		// TODO: stress test this.
		go func() {
			defer wg.Done()
			req := codec.NewRequestFromMessage(
				ctx,
				v.msg,
			)
			req.Peer = peerInfo
			s.services.ServeRPC(v, req)
		}()
	}
	if r.batch && totalRequests > 0 {
		err = doneMu.Acquire(ctx, int64(totalRequests))
		if err != nil {
			return err
		}
		err = r.mu.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		defer r.mu.Release(1)
		// write them, one by one
		_, err = r.remote.Write([]byte{'['})
		if err != nil {
			return err
		}
		for i, v := range batchResults {
			var a any
			a = v.payload
			err = r.send(ctx, &callEnv{
				v:           &a,
				err:         v.err,
				id:          v.msg.ID,
				extrafields: v.msg.ExtraFields,
			})
			if err != nil {
				return err
			}
			// write the comma or ]
			char := ','
			if i == len(batchResults)-1 {
				char = ']'
			}
			_, err = r.remote.Write([]byte{byte(char)})
			if err != nil {
				return err
			}
		}
		err = r.remote.Flush()
		if err != nil {
			return err
		}
	} else if totalRequests == 0 {
		err = r.mu.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		defer r.mu.Release(1)
		err := r.remote.Flush()
		if err != nil {
			return err
		}
	}
	wg.Wait()
	return nil
}

type callResponder struct {
	remote codec.ReaderWriter
	mu     *semaphore.Weighted

	batch        bool
	batchStarted bool
}

type callEnv struct {
	v           *any
	err         error
	id          *codec.ID
	extrafields codec.ExtraFields
}

func (c *callResponder) send(ctx context.Context, env *callEnv) (err error) {
	enc := jx.GetEncoder()
	defer jx.PutEncoder(enc)
	enc.Grow(4096)
	enc.ResetWriter(c.remote)
	enc.Obj(func(e *jx.Encoder) {
		e.Field("jsonrpc", func(e *jx.Encoder) {
			e.Str("2.0")
		})
		if env.id != nil {
			e.Field("id", func(e *jx.Encoder) {
				e.Raw(env.id.RawMessage())
			})
		}
		if env.extrafields != nil {
			for k, v := range env.extrafields {
				e.Field(k, func(e *jx.Encoder) {
					e.Raw(v)
				})
			}
		}
		if env.err != nil {
			e.Field("error", func(e *jx.Encoder) {
				codec.EncodeError(e, env.err)
			})
		} else {
			// if there is no error, we try to marshal the result
			e.Field("result", func(e *jx.Encoder) {
				if env.v != nil {
					switch cast := (*env.v).(type) {
					case json.RawMessage:
						e.Raw(cast)
					default:
						err = json.NewEncoder(e).EncodeWithOption(cast, func(eo *json.EncodeOption) {
							eo.DisableNewline = true
						})
						if err != nil {
							return
						}
					}
				} else {
					e.Null()
				}
			})
		}
	})
	// a json encoding error here is possibly fatal....
	if err != nil {
		return err
	}
	err = enc.Close()
	if err != nil {
		return err
	}
	return nil
}

type notifyEnv struct {
	method string
	dat    any
	extra  codec.ExtraFields
}

func (c *callResponder) notify(ctx context.Context, env *notifyEnv) (err error) {
	msg := &codec.Message{}
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
	enc := jx.GetEncoder()
	defer jx.PutEncoder(enc)
	enc.Grow(4096)
	enc.ResetWriter(c.remote)
	err = codec.MarshalMessage(msg, enc)
	if err != nil {
		return err
	}
	return enc.Close()
}
