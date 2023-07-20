package codec

import (
	"context"
	"io"
)

type ReaderWriter interface {
	Reader
	Writer
}

// Reader can write JSON messages to its underlying connection
// Implementations must be safe for concurrent use
type Reader interface {
	// gets the peer info
	PeerInfo() PeerInfo
	// json.RawMessage can be an array of requests. if it is, then it is a batch request
	ReadBatch(ctx context.Context) (msgs []*Message, batch bool, err error)
	// closes the connection
	Close() error
}

// Writer can write bytes messages to their underlying connection.
// Implementations must be safe for concurrent use.
type Writer interface {
	// write json blob to stream
	io.Writer
	// Flush flushes the writer to the stream  between messages
	Flush() error

	// Closed returns a channel which is closed when the connection is closed.
	Closed() <-chan struct{}
	// RemoteAddr returns the peer address of the connection.
	RemoteAddr() string
}
