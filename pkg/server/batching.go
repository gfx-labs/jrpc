package server

import (
	"context"
	"sync"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"github.com/mailgun/multibuf"
)

// serving batches is a bit complicated and we don't even use it
func serveBatch(ctx context.Context,
	incoming []*jsonrpc.Message,
	r *callResponder,
	handler jsonrpc.Handler,
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
		// create the response writer
		om, omerr := produceOutputMessage(v)
		rw := &streamingRespWriter{
			ctx:          ctx,
			sendStream:   ansBatch,
			notifyStream: r.stream,
			id:           om.ID,
			err:          omerr,
		}
		if rw.id != nil {
			totalRequests += 1
		}
		req := jsonrpc.NewRawRequest(
			ctx,
			om.ID,
			om.Method,
			om.Params,
		)
		req.Peer = r.peerinfo
		run := func() {
			defer returnWg.Done()
			handler.ServeRPC(rw, req)
			if rw.sendCalled == false && rw.id != nil {
				rw.Send(jsonrpc.Null, nil)
			}
		}
		// note that we wait for ServeRPC to return here
		// this is again, so that we can promise that batch requests, even notifications, are sequentially served.
		run()
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
