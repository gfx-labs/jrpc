package jrpc

import (
	"context"
	"io"
	"net/http"
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
	services Handler
	run      int32
	codecs   mapset.Set
	Tracing  Tracing
}

type Tracing struct {
	ErrorLogger func(remote codec.ReaderWriter, err error)
}

// NewServer creates a new server instance with no registered handlers.
func NewServer(r Handler) *Server {
	server := &Server{
		codecs: mapset.NewSet(),
		run:    1,
	}
	server.services = r
	// Register the default service providing meta information about the RPC service such
	// as the services and methods it offers.
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
		// lose
		err = remote.Close()
		if err != nil {
			s.printError(remote, err)
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
			remote.Close()
		}
	}()

	for {
		msgs, err := remote.ReadBatch(ctx)
		if err != nil {
			remote.Flush()
			s.printError(remote, err)
			return
		}
		msg, batch := codec.ParseMessage(msgs)
		env := &callEnv{
			batch: batch,
		}
		for _, v := range msg {
			rw := &callRespWriter{
				msg:           v,
				notifications: responder.toNotify,
				header:        remote.PeerInfo().HTTP.Headers,
			}
			env.responses = append(env.responses, rw)
		}
		wg := sync.WaitGroup{}
		wg.Add(len(msg))
		for _, vv := range env.responses {
			v := vv
			go func() {
				if v.msg.ID == nil {
					wg.Done()
				} else {
					defer wg.Done()
				}
				s.services.ServeRPC(v, &Request{
					ctx: ctx,
					RequestMarshaling: RequestMarshaling{
						ID:      v.msg.ID,
						Version: v.msg.Version,
						Method:  v.msg.Method,
						Params:  v.msg.Params,
						Peer:    remote.PeerInfo(),
					},
				})
			}()
		}
		wg.Wait()
		responder.toSend <- env
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
func (c *callResponder) notify(ctx context.Context, env *notifyEnv) error {
	buf := bufpool.GetStd()
	defer bufpool.PutStd(buf)
	enc := jx.GetEncoder()
	enc.ResetWriter(c.remote)
	defer jx.PutEncoder(enc)
	buf.Reset()
	enc.ObjStart()
	enc.FieldStart("jsonrpc")
	enc.Str("2.0")
	err := env.dat(buf)
	if err != nil {
		enc.FieldStart("error")
		err := codec.EncodeError(enc, err)
		if err != nil {
			return err
		}
	} else {
		enc.FieldStart("result")
		enc.Raw(buf.Bytes())
	}
	enc.ObjEnd()
	err = enc.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *callResponder) send(ctx context.Context, env *callEnv) error {
	buf := bufpool.GetStd()
	defer bufpool.PutStd(buf)
	enc := jx.GetEncoder()
	enc.Reset()
	//enc.ResetWriter(c.remote)
	defer jx.PutEncoder(enc)
	if env.batch {
		enc.ArrStart()
	}
	for _, v := range env.responses {
		if v.msg.ID == nil {
			continue
		}
		enc.ObjStart()
		enc.FieldStart("jsonrpc")
		enc.Str("2.0")
		enc.FieldStart("id")
		enc.Raw(v.msg.ID.RawMessage())
		err := v.err
		if err == nil {
			if v.dat != nil {
				buf.Reset()
				err = v.dat(buf)
				if err == nil {
					enc.FieldStart("result")
					enc.Raw(buf.Bytes())
				}
			} else {
				err = codec.NewMethodNotFoundError(v.msg.Method)
			}
		}
		if err != nil {
			enc.FieldStart("error")
			err := codec.EncodeError(enc, err)
			if err != nil {
				return err
			}
		}
		enc.ObjEnd()
	}
	if env.batch {
		enc.ArrEnd()
	}
	//err := enc.Close()
	_, err := enc.WriteTo(c.remote)
	if err != nil {
		return err
	}
	return nil
}

type callEnv struct {
	responses []*callRespWriter
	batch     bool
}

type notifyEnv struct {
	dat func(io.Writer) error
}

type callRespWriter struct {
	msg    *codec.Message
	dat    func(io.Writer) error
	err    error
	skip   bool
	header http.Header

	notifications chan *notifyEnv
}

func (c *callRespWriter) Send(v any, err error) error {
	if err != nil {
		c.err = err
		return nil
	}
	c.dat = func(w io.Writer) error {
		return json.NewEncoder(w).Encode(v)
	}
	return nil
}

func (c *callRespWriter) Option(k string, v any) {
	// no options for now
}

func (c *callRespWriter) Header() http.Header {
	return c.header
}

func (c *callRespWriter) Notify(v any) error {
	c.notifications <- &notifyEnv{
		dat: func(w io.Writer) error {
			return json.NewEncoder(w).Encode(v)
		},
	}
	return nil
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

type peerInfoContextKey struct{}

// PeerInfoFromContext returns information about the client's network connection.
// Use this with the context passed to RPC method handler functions.
//
// The zero value is returned if no connection info is present in ctx.
func PeerInfoFromContext(ctx context.Context) codec.PeerInfo {
	info, _ := ctx.Value(peerInfoContextKey{}).(codec.PeerInfo)
	return info
}
func ContextWithPeerInfo(ctx context.Context, c codec.PeerInfo) context.Context {
	return context.WithValue(ctx, peerInfoContextKey{}, c)
}
