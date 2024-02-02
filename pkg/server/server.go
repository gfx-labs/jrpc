package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/mailgun/multibuf"
	"golang.org/x/sync/errgroup"

	"gfx.cafe/open/jrpc/pkg/jjson"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/serverutil"
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
// the response back using the given codec. It will block until the codec is closed.
// the codec will return if either of these conditions are met
// 1. every request read from ReadBatch until ReadBatch returns context.Canceled is processed.
// 2. there is a server related error (failed encoding, broken conn) that was received while processing/reading messages.
func (s *Server) ServeCodec(ctx context.Context, remote jsonrpc.ReaderWriter) error {
	select {
	case <-s.lctx.Done():
		return http.ErrServerClosed
	default:
	}
	// close the remote after handling it
	defer remote.Close()
	stream := jsonrpc.NewStream(remote)
	// add a cancel to the context so we can cancel all the child tasks on return
	ctx = ContextWithPeerInfo(ctx, remote.PeerInfo())
	ctx = ContextWithMessageStream(ctx, stream)
	ctx, cn := context.WithCancel(ctx)
	defer cn()
	errCh := make(chan error, 1)
	batches := make(chan serverutil.Bundle, 1)
	go func() {
		defer func() {
			close(batches)
		}()
		for {
			// read messages from the stream synchronously
			incoming, batch, err := remote.ReadBatch(ctx)
			if err != nil {
				// if its not context canceled, aka our graceful closure, we error, otherwise we only return
				// in both cases we close the batches channel. this error will then immediately return.
				select {
				case errCh <- err:
				default:
				}
				return
			}
			select {
			case batches <- serverutil.Bundle{
				Messages: incoming,
				Batch:    batch,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()
	wg := sync.WaitGroup{}
	// this errgroup controls the max concurrent requests per codec
	egg, ctx := errgroup.WithContext(ctx)
	for batch := range batches {
		incoming, batch := batch.Messages, batch.Batch
		wg.Add(1)
		responder := &callResponder{
			peerinfo: remote.PeerInfo(),
			batch:    batch,
			stream:   stream,
		}
		egg.Go(func() error {
			return s.serve(ctx, incoming, responder)
		})
	}
	egg.Wait()
	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.cn()
	return nil
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
		ctx:          ctx,
		sendStream:   r.stream,
		notifyStream: r.stream,
	}
	om, omerr := produceOutputMessage(incoming)
	rw.id = om.ID
	rw.err = omerr
	req := jsonrpc.NewRawRequest(
		ctx,
		rw.id,
		incoming.Method,
		incoming.Params,
	)
	req.Peer = r.peerinfo
	if rw.id == nil {
		// all notification, so immediately flush a response
		err := r.stream.Flush(ctx)
		if err != nil {
			return err
		}
	}
	s.services.ServeRPC(rw, req)
	if rw.sendCalled == false && rw.id != nil {
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
		mw, err := r.stream.NewMessage(ctx)
		if err != nil {
			return err
		}
		defer mw.Close()
		if err := mw.Field("id", jsonrpc.Null); err != nil {
			return err
		}
		if err := mw.Field("error", jsonrpc.MarshalError(jsonrpc.NewInvalidRequestError("empty batch"))); err != nil {
			return err
		}
		return nil
	}

	totalRequests := 0
	// populate the envelope we are about to send. this is synchronous pre-prpcessing
	ansBuf, err := multibuf.NewWriterOnce(
		// store up to 16mb per batch in memory
		multibuf.MemBytes(16*1024*1024),
		// store up to 256gb per batch on disk
		multibuf.MaxBytes(256*1204*1024*1024),
	)
	defer ansBuf.Close()
	if err != nil {
		return err
	}
	ansStream := jsonrpc.NewStream(ansBuf)
	ansBatch, err := ansStream.NewBatch(ctx)
	if err != nil {
		return err
	}

	// create a waitgroup for when every handler returns
	returnWg := sync.WaitGroup{}
	returnWg.Add(len(incoming))
	for _, v := range incoming {
		canNext := make(chan struct{})
		// create the response writer
		rw := &streamingRespWriter{
			ctx:          ctx,
			sendStream:   ansBatch,
			notifyStream: r.stream,
		}
		om, omerr := produceOutputMessage(v)
		rw.id = om.ID
		rw.err = omerr
		if rw.id != nil {
			totalRequests += 1
			rw.done = func() {
				close(canNext)
			}
		}
		req := jsonrpc.NewRawRequest(
			ctx,
			om.ID,
			om.Method,
			om.Params,
		)
		req.Peer = r.peerinfo
		go func() {
			defer returnWg.Done()
			s.services.ServeRPC(rw, req)
			if rw.sendCalled == false && rw.id != nil {
				rw.Send(jsonrpc.Null, nil)
			}
		}()
		if rw.id != nil {
			<-canNext
		}
	}

	err = ansBatch.Close()
	if err != nil {
		return err
	}

	mr, err := ansBuf.Reader()
	if err != nil {
		return err
	}
	defer mr.Close()

	if totalRequests > 0 {
		// TODO: channel?
		err := r.stream.ReadFrom(ctx, mr)
		if err != nil {
			return err
		}
	} else if totalRequests == 0 {
		// all notification, so immediately flush, and that's the whole message
		err := r.stream.Flush(ctx)
		if err != nil {
			return err
		}
	}
	// wait for the returnWg to return
	returnWg.Wait()
	return nil
}

type callResponder struct {
	peerinfo jsonrpc.PeerInfo
	stream   *jsonrpc.MessageStream
	batch    bool
}

type callEnv struct {
	v   any
	err error
	id  *jsonrpc.ID
}

func send(env *callEnv, s *jsonrpc.MessageWriter) (err error) {
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
	defer wr.Close()
	// if is nil, just write null
	if env.v == nil {
		_, err := wr.Write(jsonrpc.Null)
		if err != nil {
			return err
		}
		return nil
	}
	// if is not nil, do switch statement
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
		err = jjson.Encode(wr, cast)
	}
	return nil
}

type notifyEnv struct {
	method string
	dat    any
}

func notify(env *notifyEnv, s *jsonrpc.MessageWriter) (err error) {
	err = s.Field("method", []byte(`"`+env.method+`"`))
	if err != nil {
		return err
	}
	// if there is no error, we try to marshal the result
	wr, err := s.Params()
	if err != nil {
		return err
	}
	// if is nil, just write null
	if env.dat == nil {
		_, err := wr.Write(jsonrpc.Null)
		if err != nil {
			return err
		}
		return nil
	}
	// if is not nil, do switch statement
	switch cast := (env.dat).(type) {
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
		err = jjson.Encode(wr, cast)
	}
	return nil
}
