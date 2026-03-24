/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package mod_json_rpc

import (
	"go.osspkg.com/errors"

	at "go.osspkg.com/goppy/v3/apigen/types"
)

type Module struct {
	FilePrefix string
}

func (Module) Name() string {
	return "json-rpc"
}

func (v Module) Build(w at.Writer, m at.GlobalMeta, files []at.File) error {
	return errors.Queue(
		func() error { return v.buildTransportModels(w, m, files) },
		func() error { return v.buildTransportHandlers(w, m, files) },
	)
}
