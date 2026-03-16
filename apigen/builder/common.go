package builder

import (
	modjsonrpc "go.osspkg.com/goppy/v3/apigen/module/mod-json-rpc"
	modparamcookie "go.osspkg.com/goppy/v3/apigen/module/mod-param-cookie"
	modparamheader "go.osspkg.com/goppy/v3/apigen/module/mod-param-header"
	"go.osspkg.com/goppy/v3/apigen/types"
)

func init() {
	types.Register[types.GlobalModule](modjsonrpc.Module{FilePrefix: "jsonrpc"})
	types.Register[types.ParamModule](modparamcookie.Module{})
	types.Register[types.ParamModule](modparamheader.Module{})
}

const (
	TagModule = "module"
)
