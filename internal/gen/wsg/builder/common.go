package builder

import (
	modjsonrpc "go.osspkg.com/goppy/v3/internal/gen/wsg/module/mod-json-rpc"
	"go.osspkg.com/goppy/v3/internal/gen/wsg/types"
)

func init() {
	types.Register(modjsonrpc.JSONRPCTransport{FilePrefix: "jsonrpc"})
	//types.Register(HTTPTransport{})
}

const (
	TagModule = "module"
)
