package server

import (
	"bytes"
	"context"
	"sync"

	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
	"github.com/mailgun/multibuf"
	"github.com/valyala/bytebufferpool"
)

func (s *Server) serveBatch(ctx context.Context,
	incoming []*jsonrpc.Message,
	r *callResponder,
	handler jsonrpc.Handler,
) error {
	if s.BatchLimit > 0 && len(incoming) > s.BatchLimit {
		// if the batch is too large, send an immediate error as well
		mw, err := r.stream.NewMessage(ctx)
		if err != nil {
			return err
		}
		defer mw.Close()
		if err := mw.Field("id", jsonrpc.Null); err != nil {
			return err
		}
		if err := mw.Field("error", jsonrpc.MarshalError(jsonrpc.NewInvalidRequestError("batch too large"))); err != nil {
			return err
		}
		return nil
	}
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
	if s.BatchParallel {
		return s.serveBatchParallel(ctx, incoming, r, handler)
	} else {
		return s.serveBatchSerial(ctx, incoming, r, handler)
	}
}

// a less memory efficient but faster for high RTT connections, like the REST transport, due to inability to respond a single request at a time
func (s *Server) serveBatchParallel(ctx context.Context,
	incoming []*jsonrpc.Message,
	r *callResponder,
	handler jsonrpc.Handler,
) error {
	// run them all in parallel, and collect the responses in buffer pools
	returnWg := sync.WaitGroup{}
	bufs := make([]*bytebufferpool.ByteBuffer, 0, len(incoming))
	totalRequests := 0
	for _, v := range incoming {
		buf := bytebufferpool.Get()
		bufs = append(bufs, buf)
		defer bytebufferpool.Put(buf)
		stream := jsonrpc.NewStream(buf)
		om, omerr := s.produceOutputMessage(v)
		returnWg.Add(1)
		req := jsonrpc.NewRawRequest(
			ctx,
			om.ID,
			om.Method,
			om.Params,
		)
		req.Peer = r.peerinfo
		rw := &streamingRespWriter{
			ctx:          ctx,
			sendStream:   stream,
			notifyStream: stream,
			id:           om.ID,
			err:          omerr,
		}
		if rw.id != nil {
			totalRequests = totalRequests + 1
		}
		go func() {
			defer returnWg.Done()
			handler.ServeRPC(rw, req)
			if rw.sendCalled == false && rw.id != nil {
				rw.Send(jsonrpc.Null, nil)
			}
		}()
	}
	returnWg.Wait()

	if totalRequests > 0 {
		bufw, err := r.stream.NewBatch(ctx)
		if err != nil {
			return err
		}
		for _, buf := range bufs {
			err := bufw.WriteMessage(ctx, bytes.TrimSpace(buf.B))
			if err != nil {
				return err
			}
		}
		err = bufw.Close()
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

	return nil
}

// a "more" memory efficient implementation of batching for sequential processing of large data.
func (s *Server) serveBatchSerial(ctx context.Context,
	incoming []*jsonrpc.Message,
	r *callResponder,
	handler jsonrpc.Handler,
) error {

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
		om, omerr := s.produceOutputMessage(v)
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
		err = r.stream.Flush(ctx)
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
