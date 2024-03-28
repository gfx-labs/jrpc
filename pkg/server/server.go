package server

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/serverutil"
)

// ServeCodec reads incoming requests from codec, calls the appropriate callback and writes
// the response back using the given codec. It will block until the codec is closed.
// the codec will return if either of these conditions are met
// 1. every request read from ReadBatch until ReadBatch returns context.Canceled is processed.
// 2. there is a server related error (failed encoding, broken conn) that was received while processing/reading messages.
func ServeCodec(ctx context.Context, remote jsonrpc.ReaderWriter, handler jsonrpc.Handler) error {
	// close the remote after handling it
	defer remote.Close()
	stream := jsonrpc.NewStream(remote)
	// add a cancel to the context so we can cancel all the child tasks on return
	ctx = ContextWithMessageStream(ContextWithPeerInfo(
		ctx,
		remote.PeerInfo(),
	), stream,
	)
	egg, ctx := errgroup.WithContext(ctx)
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
				if errors.Is(err, jsonrpc.ErrNoMoreBatches) || errors.Is(err, context.Canceled) {
					return
				}
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
	// this errgroup controls the max concurrent requests per codec
	for batch := range batches {
		incoming, batch := batch.Messages, batch.Batch
		responder := &callResponder{
			peerinfo: remote.PeerInfo(),
			batch:    batch,
			stream:   stream,
		}
		egg.Go(func() error {
			return serve(ctx, incoming, responder, handler)
		})
	}
	err := egg.Wait()
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

type callResponder struct {
	peerinfo jsonrpc.PeerInfo
	stream   *jsonrpc.MessageStream
	batch    bool
}

func serve(ctx context.Context,
	incoming []*jsonrpc.Message,
	r *callResponder,
	handler jsonrpc.Handler,
) error {
	if r.batch {
		return serveBatch(ctx, incoming, r, handler)
	} else {
		return serveSingle(ctx, incoming[0], r, handler)
	}
}

func serveSingle(ctx context.Context,
	incoming *jsonrpc.Message,
	r *callResponder,
	handler jsonrpc.Handler,
) error {
	om, omerr := produceOutputMessage(incoming)
	rw := &streamingRespWriter{
		ctx:          ctx,
		sendStream:   r.stream,
		notifyStream: r.stream,
		id:           om.ID,
		err:          omerr,
	}
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
	handler.ServeRPC(rw, req)
	if rw.sendCalled == false {
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
	// NOTE: in the past, a zero length method was an invalid request
	// now that is no longer the case
	//// zero length method is always invalid request
	if len(out.Method) == 0 {
		// assume if the method is not there AND the id is not there that it's a REQUEST not notification
		// this makes sure we add 1 to totalRequests
		if out.ID == nil {
			out.ID = jsonrpc.NewNullIDPtr()
		}
		//	err = jsonrpc.NewInvalidRequestError("invalid request")
	}

	return
}
