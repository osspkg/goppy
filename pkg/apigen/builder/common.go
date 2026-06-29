/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package builder

import (
	modjsonrpcclient "go.osspkg.com/goppy/v3/pkg/apigen/module/mod-json-rpc-client"
	modjsonrpcserver "go.osspkg.com/goppy/v3/pkg/apigen/module/mod-json-rpc-server"
	modparamcookie "go.osspkg.com/goppy/v3/pkg/apigen/module/mod-param-cookie"
	modparamheader "go.osspkg.com/goppy/v3/pkg/apigen/module/mod-param-header"
	modvalidate "go.osspkg.com/goppy/v3/pkg/apigen/module/mod-validate"
	"go.osspkg.com/goppy/v3/pkg/apigen/types"
)

func init() {
	types.Register[types.GlobalModule](modjsonrpcserver.Module{FilePrefix: "jsonrpc_server"})
	types.Register[types.GlobalModule](modjsonrpcclient.Module{FilePrefix: "jsonrpc_client"})
	types.Register[types.ParamModule](modparamcookie.Module{})
	types.Register[types.ParamModule](modparamheader.Module{})
	types.Register[types.ParamModule](modvalidate.Module{})
}
