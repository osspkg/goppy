/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

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
