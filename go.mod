module gfx.cafe/open/jrpc

go 1.21

replace github.com/goccy/go-json v0.10.2 => github.com/elee1766/go-json v0.10.2-1

replace github.com/go-faster/jx v1.1.0 => ../../../github.com/elee1766/jx

require (
	gfx.cafe/open/websocket v1.9.2
	gfx.cafe/util/go/bufpool v0.0.0-20230721185457-c559e86c829c
	gfx.cafe/util/go/frand v0.0.0-20230721185457-c559e86c829c
	gfx.cafe/util/go/generic v0.0.0-20230721185457-c559e86c829c
	github.com/alecthomas/kong v0.8.0
	github.com/go-faster/jx v1.1.0
	github.com/goccy/go-json v0.10.2
	github.com/rs/xid v1.5.0
	github.com/stretchr/testify v1.8.4
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/aead/chacha20 v0.0.0-20180709150244-8b13a72661da // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-faster/errors v0.6.1 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	golang.org/x/exp v0.0.0-20230206171751-46f607a40771 // indirect
	golang.org/x/sys v0.12.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
