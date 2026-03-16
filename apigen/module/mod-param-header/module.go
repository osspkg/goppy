/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package mod_param_header

import (
	"fmt"

	. "go.osspkg.com/gogen/golang" //nolint:staticcheck

	at "go.osspkg.com/goppy/v3/apigen/types"
)

type Module struct{}

func (Module) Name() string {
	return "header"
}

func (v Module) Build(w at.Joiner, m at.ParamMeta, value at.Param) error {
	switch m.Type {
	case at.ParamIn:
		return v.generateIn(w, m, value)
	case at.ParamOut:
		return v.generateOut(w, m, value)
	default:
		return fmt.Errorf("unknown type")
	}
}

func (v Module) generateIn(w at.Joiner, m at.ParamMeta, p at.Param) error {
	m.Import.Set("web", "go.osspkg.com/goppy/v3/web")

	w.Join(
		List(ID(m.CodeName), Raw("_")).Op("=").
			ID("web").Op(".").ID("StrTo").Raw("[").Pkg(p.Pkg).ID(p.Type).Raw("]").
			Call(
				ID("webCtx").Op(".").ID("Header").Call().Op(".").ID("Get").Call(Text(m.Value)),
			),
		If().ID("err").Op("!=").Nil().Block(
			ID("err").Op("=").Pkg("fmt").ID("Errorf").Bracket(
				Text("invalid request: %w"),
				ID("err"),
			),
			Return().List(
				Nil(),
				ID("err"),
			)),
	)

	return nil
}

func (v Module) generateOut(w at.Joiner, m at.ParamMeta, _ at.Param) error {
	m.Import.Set("fmt", "fmt")

	w.Join(
		If().ID("err").Op("!=").Nil().Block(Return()),
		ID("webCtx").Op(".").ID("Header").Call().Op(".").ID("Set").Bracket(
			Text(m.Value),
			Pkg("fmt").ID("Sprintf").Bracket(
				List(Text("%v"), Raw(m.CodeName)),
			),
		),
	)
	return nil
}
