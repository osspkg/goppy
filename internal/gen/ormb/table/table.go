/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package table

import (
	"bytes"
	"fmt"
)

type Table struct {
	ModelName string
	TableName string
	Fields    []TField
	_attrs    *Attrs
}

func CreateTableType() *Table {
	return &Table{
		ModelName: "",
		TableName: "",
		Fields:    []TField{},
		_attrs:    NewAttrs(),
	}
}

func (t *Table) Attrs() *Attrs {
	return t._attrs
}

func (t *Table) GetAttrsByKey(key AttrKeyType) ([]Attr, bool) {
	result, _ := t._attrs.GetByKey(key)
	for _, field := range t.Fields {
		vals, ok := field.Attrs().GetByKey(key)
		if ok {
			result = append(result, vals...)
		}
	}
	return result, len(result) > 0
}

func (t *Table) GetAttrsByKeyDo(key AttrKeyType, do AttrDoType) ([]Attr, bool) {
	result, _ := t._attrs.GetByKeyDo(key, do)
	for _, field := range t.Fields {
		vals, ok := field.Attrs().GetByKeyDo(key, do)
		if ok {
			result = append(result, vals...)
		}
	}
	return result, len(result) > 0
}

func (t *Table) String() string {
	buf := bytes.NewBufferString("")

	fmt.Fprintf(buf, "\tModelName=%s\n", t.ModelName)
	fmt.Fprintf(buf, "\tTableName=%s\n", t.TableName)

	for i, datum := range t.Fields {
		fmt.Fprintf(buf, "\tField[%d]=%s\n", i, datum.String())
	}

	for i, datum := range t._attrs.data {
		fmt.Fprintf(buf, "\t- attr[%d]={key:'%s',do:'%v',val:'%v'};\n", i, datum.Key, datum.Do, datum.Value)
	}

	return buf.String()
}
