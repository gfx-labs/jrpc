package jmux

import (
	"gfx.cafe/open/jrpc/pkg/codec"
)

// ChainHandler is a Handler with support for handler composition and
// execution.
type ChainHandler = codec.ChainHandler

// Chain returns a Middlewares type from a slice of middleware handlers.
var Chain = codec.Chain
