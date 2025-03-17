/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package visitor

import (
	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/fields"
)

const (
	tag = "//gen:orm"
)

type Model struct {
	Name   fields.ModelNameType
	Table  fields.TableType
	Fields []fields.TField
	Attr   *fields.Attrs
}

func NewModel() *Model {
	return &Model{
		Fields: make([]fields.TField, 0),
		Attr:   fields.NewAttrs(),
	}
}
