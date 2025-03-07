/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package fields

import (
	"fmt"
	"reflect"
	"strings"
)

type AttrType int8

const (
	AttrIndexPK   AttrType = 1
	AttrIndexFK   AttrType = 2
	AttrIndexUniq AttrType = 3
	AttrTableName AttrType = 4
	AttrAction    AttrType = 5
	AttrFieldLink AttrType = 6
	AttrFieldLen  AttrType = 7
	AttrFieldCol  AttrType = 8
	AttrFieldAuto AttrType = 9
)

const (
	AttrValueActionRO = "ro"
)

type Attr struct {
	Type    AttrType
	Key     string
	Value   []string
	notUniq bool
}

func ParseAttr(v string) (*Attr, bool) {
	vv := strings.Split(v, "=")
	if len(vv) != 2 {
		return nil, false
	}

	attrType := vv[0]
	attrKey := ""
	attrVal := make([]string, 0)

	vvv := strings.Split(vv[1], ":")

	switch len(vvv) {
	case 1:
		attrKey = vvv[0]
		attrVal = append(attrVal, vvv[0])
	case 2:
		attrKey = vvv[0]
		attrVal = append(attrVal, strings.Split(vvv[1], ",")...)
	default:
		panic(fmt.Sprintf("invalid attr: %s, must be format: <type>=<key>:<value>,<value>,...", v))
	}

	switch attrType {
	case "table":
		return &Attr{Type: AttrTableName, Value: attrVal}, true
	case "action":
		return &Attr{Type: AttrAction, Value: attrVal}, true
	case "col":
		return &Attr{Type: AttrFieldCol, Value: attrVal}, true
	case "len":
		return &Attr{Type: AttrFieldLen, Value: attrVal}, true
	case "auto":
		return &Attr{Type: AttrFieldAuto, Key: attrKey, Value: attrVal, notUniq: true}, true
	case "link":
		return &Attr{Type: AttrFieldLink, Key: attrKey, Value: attrVal}, true
	case "index":

		switch attrKey {
		case "fk":
			return &Attr{Type: AttrIndexFK, Value: attrVal, notUniq: true}, true
		case "pk":
			return &Attr{Type: AttrIndexPK, notUniq: true}, true
		case "uniq":
			return &Attr{Type: AttrIndexUniq, Value: attrVal, notUniq: true}, true
		default:
			panic(fmt.Sprintf("unknow index: %s, must be pk,fk,uniq", v))
		}

	default:
		return nil, false
	}
}

type Attrs struct {
	data []Attr
}

func NewAttrs() *Attrs {
	return &Attrs{data: make([]Attr, 0)}
}

func (v *Attrs) Set(a Attr) {
	aa, ok := v.Get(a.Type)
	if ok {
		for _, _a := range aa {
			if _a.notUniq && _a.Key == a.Key && reflect.DeepEqual(_a.Value, a.Value) {
				return
			}
		}
	}
	v.data = append(v.data, a)
}

func (v *Attrs) FirstValue(t AttrType) (string, bool) {
	for _, datum := range v.data {
		if datum.Type == t {
			if len(datum.Value) > 0 {
				return datum.Value[0], true
			}
			return "", false
		}
	}
	return "", false
}

func (v *Attrs) Get(t AttrType) (a []Attr, ok bool) {
	for _, datum := range v.data {
		if datum.Type == t {
			a = append(a, datum)
			ok = true
		}
	}
	return
}

func (v *Attrs) GetOne(t AttrType) (Attr, bool) {
	for _, datum := range v.data {
		if datum.Type == t {
			return datum, true
		}
	}
	return Attr{}, false
}

func (v *Attrs) All() []Attr {
	return v.data
}
