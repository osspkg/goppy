/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package table

import (
	"bytes"
	"fmt"
)

type FieldType int8

const (
	FieldTypeSingle FieldType = 1
	FieldTypeArray  FieldType = 2
	FieldTypeLink   FieldType = 3
)

type TField interface {
	Name() string
	Type() FieldType
	GoType() string
	Attrs() *Attrs
	String() string
}

type base struct {
	name      string
	fieldType FieldType
	goType    string
	attr      *Attrs
}

func (b base) Name() string {
	return b.name
}

func (b base) Type() FieldType {
	return b.fieldType
}

func (b base) GoType() string {
	return b.goType
}

func (b base) Attrs() *Attrs {
	return b.attr
}

func (b base) String() string {
	buf := bytes.NewBufferString("")

	fmt.Fprintf(buf, "name=%s; ", b.Name())
	fmt.Fprintf(buf, "type=%#v; ", b.Type())
	fmt.Fprintf(buf, "go-type=%s; ", b.GoType())

	for i, datum := range b.attr.data {
		fmt.Fprintf(buf, "attr[%d]={key:'%s',do:'%v',val:'%v'}; ", i, datum.Key, datum.Do, datum.Value)
	}

	return buf.String()
}
