package server

import (
	"errors"
	"fmt"
	"io"
)

var _ io.Writer = (*Writer)(nil)
var ErrThresholdExceeded = errors.New("stream size exceeds threshold")

// Writer wraps w with writing length limit.
//
// To create Writer, use NewWriter().
type Writer struct {
	w                    io.Writer
	written              int
	limit                int
	regardOverSizeNormal bool
}

// NewWriter create a writer that writes at most n bytes.
//
// regardOverSizeNormal controls whether Writer.Write() returns error
// when writing totally more bytes than n, or do no-op to inner w,
// pretending writing is processed normally.
func newWriter(w io.Writer, n int, regardOverSizeNormal bool) *Writer {
	return &Writer{
		w:                    w,
		written:              0,
		limit:                n,
		regardOverSizeNormal: regardOverSizeNormal,
	}
}

// Writer implements io.Writer
func (lw *Writer) Write(p []byte) (n int, err error) {
	if lw.written >= lw.limit {
		if lw.regardOverSizeNormal {
			n = len(p)
			lw.written += n
			return
		}

		err = fmt.Errorf("threshold is %d bytes: %w", lw.limit, ErrThresholdExceeded)
		return
	}

	var (
		overSized   bool
		originalLen int
	)

	left := lw.limit - lw.written
	if originalLen = len(p); originalLen > left {
		overSized = true
		p = p[0:left]
	}
	n, err = lw.w.Write(p)
	lw.written += n
	if overSized && err == nil {
		// Write must return a non-nil error if it returns n < len(p).
		if lw.regardOverSizeNormal {
			return originalLen, nil
		}

		err = fmt.Errorf("threshold is %d bytes: %w", lw.limit, ErrThresholdExceeded)
		return
	}

	return
}
