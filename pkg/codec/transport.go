package codec

import (
	"context"
	"io"
	"net"
)

type Listener interface {
	Accept() (ReaderWriter, error)
	Close() error
	Addr() net.Addr
}

// ReaderWriter represents a single stream
// this stream can be used to send/receive an arbitrary amount of requests and notifications
type ReaderWriter interface {
	Reader
	Writer
}

// Reader can write JSON messages to its underlying connection
// Implementations must be safe for concurrent use
type Reader interface {
	// gets the peer info
	PeerInfo() PeerInfo
	// reads a batch of messages
	ReadBatch(ctx context.Context) (msgs []*Message, batch bool, err error)
	// closes the connection
	Close() error
}

// Writer can write bytes messages to their underlying connection.
// Implementations must be safe for concurrent use.
type Writer interface {
	// write json blob to stream
	io.Writer
	Flush() error
	// Closed returns a channel which is closed when the connection is closed.
	Closed() <-chan struct{}
}
