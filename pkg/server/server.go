package server

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"golang.org/x/sync/semaphore"

	"gfx.cafe/open/jrpc/pkg/jjson"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"

	"github.com/go-faster/jx"
)

// Server is an RPC server.
// it is in charge of calling the handler on the message object, the json encoding of responses, and dealing with batch semantics.
// a server can be used to listenandserve multiple codecs at a time
type Server struct {
	services jsonrpc.Handler

	lctx context.Context
	cn   context.CancelFunc
}

// NewServer creates a new server instance with no registered handlers.
func NewServer(r jsonrpc.Handler) *Server {
	server := &Server{services: r}
	server.lctx, server.cn = context.WithCancel(context.Background())
	return server
}

// ServeCodec reads incoming requests from codec, calls the appropriate callback and writes
// the response back using the given codec. It will block until the codec is closed
func (s *Server) ServeCodec(ctx context.Context, remote jsonrpc.ReaderWriter) error {
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
				err = s.serve(ctx, incoming, responder)
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

func (s *Server) serve(ctx context.Context,
	incoming []*jsonrpc.Message,
	r *callResponder,
) error {
	if r.batch {
		return s.serveBatch(ctx, incoming, r)
	} else {
		return s.serveSingle(ctx, incoming[0], r)
	}
}

func (s *Server) serveSingle(ctx context.Context,
	incoming *jsonrpc.Message,
	r *callResponder,
) error {
	rw := &streamingRespWriter{
		ctx: ctx,
		cr:  r,
	}
	rw.msg, rw.err = produceOutputMessage(incoming)
	req := jsonrpc.NewRawRequest(
		ctx,
		rw.msg.ID,
		rw.msg.Method,
		rw.msg.Params,
	)
	req.Peer = r.remote.PeerInfo()
	if rw.msg.ID == nil {
		// all notification, so immediately flush
		err := r.mu.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		defer r.mu.Release(1)
		err = r.remote.Flush()
		if err != nil {
			return err
		}
	}
	s.services.ServeRPC(rw, req)
	if rw.sendCalled == false && rw.msg.ID != nil {
		rw.Send(jsonrpc.Null, nil)
	}
	return nil
}

func produceOutputMessage(inputMessage *jsonrpc.Message) (out *jsonrpc.Message, err error) {
	// a nil incoming message means return an invalid request.
	if inputMessage == nil {
		inputMessage = &jsonrpc.Message{ID: jsonrpc.NewNullIDPtr()}
		err = jsonrpc.NewInvalidRequestError("invalid request")
	}
	out = inputMessage
	out.Error = nil
	// zero length method is always invalid request
	if len(out.Method) == 0 {
		// assume if the method is not there AND the id is not there that it's an invalid REQUEST not notification
		// this makes sure we add 1 to totalRequests
		if out.ID == nil {
			out.ID = jsonrpc.NewNullIDPtr()
		}
		err = jsonrpc.NewInvalidRequestError("invalid request")
	}

	return
}

func (s *Server) serveBatch(ctx context.Context,
	incoming []*jsonrpc.Message,
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
			id:  jsonrpc.NewNullIDPtr(),
			err: jsonrpc.NewInvalidRequestError("empty batch"),
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

	rs := []*batchingRespWriter{}

	totalRequests := 0
	// populate the envelope we are about to send. this is synchronous pre-prpcessing
	for _, v := range incoming {
		// create the response writer
		rw := &batchingRespWriter{
			ctx: ctx,
			cr:  r,
		}
		rs = append(rs, rw)
		rw.msg, rw.err = produceOutputMessage(v)
		// requests and malformed requests both count as requests
		if rw.msg.ID != nil {
			totalRequests += 1
		}
	}

	// create a waitgroup for when every handler returns
	returnWg := sync.WaitGroup{}
	returnWg.Add(len(rs))
	// for each item in the envelope
	peerInfo := r.remote.PeerInfo()
	batchResults := []*batchingRespWriter{}

	respWg := &sync.WaitGroup{}
	respWg.Add(totalRequests)

	for _, vRef := range rs {
		v := vRef
		if v.msg.ID != nil {
			v.wg = respWg
			batchResults = append(batchResults, v)
		}
		// now process each request in its own goroutine
		// TODO: stress test this.
		go func() {
			defer returnWg.Done()
			req := jsonrpc.NewRawRequest(
				ctx,
				v.msg.ID,
				v.msg.Method,
				v.msg.Params,
			)
			req.Peer = peerInfo
			s.services.ServeRPC(v, req)
			if v.sendCalled == false && v.err == nil {
				v.Send(jsonrpc.Null, nil)
			}
		}()
	}

	if totalRequests > 0 {
		// TODO: channel?
		respWg.Wait()
		err := r.mu.Acquire(ctx, 1)
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
			err = r.send(ctx, &callEnv{
				v:   v.payload,
				err: v.err,
				id:  v.msg.ID,
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
		// all notification, so immediately flush
		err := r.mu.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		defer r.mu.Release(1)
		err = r.remote.Flush()
		if err != nil {
			return err
		}
	}
	returnWg.Wait()
	return nil
}

type callResponder struct {
	remote jsonrpc.ReaderWriter
	mu     *semaphore.Weighted

	batch        bool
	batchStarted bool
}

type callEnv struct {
	v   any
	err error
	id  *jsonrpc.ID
}

func (c *callResponder) send(ctx context.Context, env *callEnv) (err error) {
	w := c.remote
	s, err := jsonrpc.NewStream(w).NewMessage()
	if err != nil {
		return err
	}
	defer s.Close()
	if env.id != nil {
		s.Field("id", env.id.RawMessage())
	}
	if env.err != nil {
		s.Field("error", jsonrpc.MarshalError(env.err))
		return nil
	}
	// if there is no error, we try to marshal the result
	wr, err := s.Result()
	if err != nil {
		return err
	}
	if env.v == nil {
		_, err := wr.Write(jsonrpc.Null)
		if err != nil {
			return err
		}
		return nil
	}
	switch cast := (env.v).(type) {
	case json.RawMessage:
		if len(cast) == 0 {
			_, err := wr.Write(jsonrpc.Null)
			if err != nil {
				return err
			}
		} else {
			_, err := wr.Write(cast)
			if err != nil {
				return err
			}
		}
	default:
		err = jjson.Encode(w, cast)
	}
	return nil
}

type notifyEnv struct {
	method string
	dat    any
}

func (c *callResponder) notify(ctx context.Context, env *notifyEnv) (err error) {
	msg := &jsonrpc.Message{}
	//  allocate a temp buffer for this packet
	buf := jjson.GetBuf()
	defer jjson.PutBuf(buf)
	err = jjson.Encode(buf, env.dat)
	if err != nil {
		msg.Error = err
	} else {
		msg.Params = buf.Bytes()
	}
	// add the method
	msg.Method = env.method
	enc := jx.GetEncoder()
	defer jx.PutEncoder(enc)
	enc.Grow(4096)
	enc.ResetWriter(c.remote)
	err = jsonrpc.MarshalMessage(msg, enc)
	if err != nil {
		return err
	}
	return enc.Close()
}
