package builder

import (
	modjsonrpc "go.osspkg.com/goppy/v3/apigen/module/mod-json-rpc"
	"go.osspkg.com/goppy/v3/apigen/types"
)

func init() {
	types.Register[types.GlobalModule](modjsonrpc.JSONRPCTransport{FilePrefix: "jsonrpc"})
}

const (
	TagModule = "module"
)
