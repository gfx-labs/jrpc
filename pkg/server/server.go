package server

import (
	"bytes"
	"context"
	"sync"
	"sync/atomic"

	"gfx.cafe/open/jrpc/pkg/codec"

	"gfx.cafe/util/go/bufpool"

	mapset "github.com/deckarep/golang-set"
	"github.com/go-faster/jx"
	"github.com/goccy/go-json"
)

// Server is an RPC server.
type Server struct {
	services codec.Handler
	run      int32
	codecs   mapset.Set
	Tracing  Tracing
}

type Tracing struct {
	ErrorLogger func(remote codec.ReaderWriter, err error)
}

// NewServer creates a new server instance with no registered handlers.
func NewServer(r codec.Handler) *Server {
	server := &Server{
		codecs: mapset.NewSet(),
		run:    1,
	}
	server.services = r
	return server
}

func (s *Server) printError(remote codec.ReaderWriter, err error) {
	if err != nil {
		return
	}
	if s.Tracing.ErrorLogger != nil {
		s.Tracing.ErrorLogger(remote, err)
	}
}

func (s *Server) codecLoop(ctx context.Context, remote codec.ReaderWriter, responder *callResponder) error {
	incoming, batch, err := remote.ReadBatch(ctx)
	if err != nil {
		remote.Flush()
		s.printError(remote, err)
		return err
	}
	env := &callEnv{
		batch: batch,
	}

	// check for empty batch
	if batch && len(incoming) == 0 {
		// if it is empty batch, send the empty batch warning
		responder.toSend <- &callEnv{
			responses: []*callRespWriter{{
				pkt: &codec.Message{
					ID:    codec.NewNullIDPtr(),
					Error: codec.NewInvalidRequestError("empty batch"),
				},
			}},
			batch: false,
		}
		return nil
	}

	// populate the envelope
	for _, v := range incoming {
		rw := &callRespWriter{
			pkt: &codec.Message{
				ID: codec.NewNullIDPtr(),
			},
			msg: &codec.Message{
				ID: codec.NewNullIDPtr(),
			},
			notifications: responder.toNotify,
			header:        remote.PeerInfo().HTTP.Headers,
		}
		if v != nil {
			rw.msg = v
			if v.ID != nil {
				rw.pkt.ID = v.ID
			}
		}
		env.responses = append(env.responses, rw)
	}

	// create a waitgroup
	wg := sync.WaitGroup{}
	wg.Add(len(env.responses))
	for _, vRef := range env.responses {
		v := vRef
		// early respond to nil requests
		if v.msg == nil || len(v.msg.Method) == 0 {
			v.pkt.Error = codec.NewInvalidRequestError("invalid request")
			wg.Done()
			continue
		}
		if v.msg.ID == nil || v.msg.ID.IsNull() {
			// it's a notification, so we mark skip and we don't write anything for it
			v.skip = true
			wg.Done()
			continue
		}
		go func() {
			defer wg.Done()
			r := codec.NewRequestFromMessage(
				ctx,
				v.msg,
			)
			r.Peer = remote.PeerInfo()
			s.services.ServeRPC(v, r)
		}()
	}
	wg.Wait()
	responder.toSend <- env
	return nil
}

// ServeCodec reads incoming requests from codec, calls the appropriate callback and writes
// the response back using the given codec. It will block until the codec is closed or the
// server is stopped. In either case the codec is closed.
func (s *Server) ServeCodec(pctx context.Context, remote codec.ReaderWriter) {
	defer remote.Close()

	// Don't serve if server is stopped.
	if atomic.LoadInt32(&s.run) == 0 {
		return
	}
	// Add the codec to the set so it can be closed by Stop.
	s.codecs.Add(remote)
	defer s.codecs.Remove(remote)

	responder := &callResponder{
		toSend:   make(chan *callEnv, 8),
		toNotify: make(chan *notifyEnv, 8),
		remote:   remote,
	}

	ctx, cn := context.WithCancel(pctx)
	defer cn()
	ctx = ContextWithPeerInfo(ctx, remote.PeerInfo())
	go func() {
		defer cn()
		err := responder.run(ctx)
		if err != nil {
			s.printError(remote, err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			remote.Close()
		default:
		}
		err := s.codecLoop(ctx, remote, responder)
		if err != nil {
			s.printError(remote, err)
			return
		}
	}
}

// Stop stops reading new requests, waits for stopPendingRequestTimeout to allow pending
// requests to finish, then closes all codecs which will cancel pending requests and
// subscriptions.
func (s *Server) Stop() {
	if atomic.CompareAndSwapInt32(&s.run, 1, 0) {
		s.codecs.Each(func(c any) bool {
			c.(codec.ReaderWriter).Close()
			return true
		})
	}
}

type callResponder struct {
	toSend   chan *callEnv
	toNotify chan *notifyEnv
	remote   codec.ReaderWriter
}

func (c *callResponder) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case env := <-c.toSend:
			err := c.send(ctx, env)
			if err != nil {
				return err
			}
		case env := <-c.toNotify:
			err := c.notify(ctx, env)
			if err != nil {
				return err
			}
		}
		if c.remote != nil {
			c.remote.Flush()
		}
	}
}

type notifyEnv struct {
	method string
	dat    any
	extra  []codec.RequestField
}

func (c *callResponder) notify(ctx context.Context, env *notifyEnv) error {
	enc := jx.GetEncoder()
	enc.Grow(4096)
	enc.ResetWriter(c.remote)
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
	return enc.Close()
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
			return nil
		}
	}
	// create the streaming encoder
	enc := jx.GetEncoder()
	enc.Grow(4096)
	enc.ResetWriter(c.remote)
	defer jx.PutEncoder(enc)
	if env.batch {
		enc.ArrStart()
	}
	for _, v := range env.responses {
		msg := v.pkt
		// if we are a batch AND we are supposed to skip, then continue
		// this means that for a non-batch notification, we do not skip!
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
	err = enc.Close()
	if err != nil {
		return err
	}
	return nil
}
