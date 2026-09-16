package jmux

import (
	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
)

// ChainHandler is a Handler with support for handler composition and
// execution.
type ChainHandler = jsonrpc.ChainHandler

// Chain returns a Middlewares type from a slice of middleware handlers.
var Chain = jsonrpc.Chain
