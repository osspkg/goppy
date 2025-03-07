/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package fields

type ModelNameType string

type TableType string

type FieldType int8

const (
	FieldTypeSingle FieldType = 1
	FieldTypeArray  FieldType = 2
	FieldTypeLink   FieldType = 3
)

type TField interface {
	Name() string
	Col() string
	Type() FieldType
	RawType() string
	Attr() *Attrs
}

type base struct {
	name      string
	col       string
	fieldType FieldType
	rawType   string
	attr      *Attrs
}

func (b base) Name() string {
	return b.name
}

func (b base) Col() string {
	return b.col
}

func (b base) Type() FieldType {
	return b.fieldType
}

func (b base) RawType() string {
	return b.rawType
}

func (b base) Attr() *Attrs {
	return b.attr
}
