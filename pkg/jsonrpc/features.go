package jsonrpc

type Hijacker interface {
	Hijack() (send MessageStreamer, notify MessageStreamer, err error)
}
