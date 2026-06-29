/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package mod_validate

import (
	"fmt"
	"strings"

	"go.osspkg.com/do"
	. "go.osspkg.com/gogen/golang" //nolint:staticcheck
	"go.osspkg.com/gogen/types"

	at "go.osspkg.com/goppy/v3/pkg/apigen/types"
)

type Module struct{}

func (Module) Name() string {
	return "validate"
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

func (v Module) generateIn(w at.Joiner, m at.ParamMeta, _ at.Param) error {
	m.Import.Set("validate", "go.osspkg.com/validate")

	var handleSrc []types.Token
	validTags := strings.Split(m.Value, ",")
	require := do.Include(validTags, "required")
	for _, tag := range validTags {
		if tag == "required" {
			continue
		}
		if require {
			handleSrc = append(handleSrc,
				ID("cb").Op(".").ID("Require").Call(Text(tag), Raw(m.CodeName)),
			)
		} else {
			handleSrc = append(handleSrc,
				ID("cb").Op(".").ID("Optional").Call(Text(tag), Raw(m.CodeName)),
			)
		}
	}

	w.Join(
		List(ID("err")).Op("=").
			ID("validate").Op(".").ID("Global").Bracket().Op(".").
			ID("Validate").Bracket(
			ID("ctx"),
			Func().Bracket(
				ID("cb").Pkg("validate").ID("Callback"),
			).Block(handleSrc...),
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

func (v Module) generateOut(_ at.Joiner, _ at.ParamMeta, _ at.Param) error {
	return nil
}
