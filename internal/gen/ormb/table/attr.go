/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package table

import (
	"fmt"
	"strings"
)

type (
	AttrKeyType string
	AttrDoType  string
)

const (
	AttrKeyTableName AttrKeyType = "table"
	AttrKeyIndex     AttrKeyType = "index"
	AttrKeyCRUD      AttrKeyType = "crud"
	AttrKeyFieldCol  AttrKeyType = "col"
	AttrKeyFieldLen  AttrKeyType = "len"
	AttrKeyFieldAuto AttrKeyType = "auto"
)

const (
	AttrDoEmpty     AttrDoType = ""
	AttrDoCreate    AttrDoType = "c"
	AttrDoUpdate    AttrDoType = "u"
	AttrDoIndexPK   AttrDoType = "pk"
	AttrDoIndexFK   AttrDoType = "fk"
	AttrDoIndexUniq AttrDoType = "unq"
	AttrDoIndexIdx  AttrDoType = "idx"
)

const (
	AttrValueCRUDc = "c"
	AttrValueCRUDr = "r"
	AttrValueCRUDu = "u"
	AttrValueCRUDd = "d"
)

func GetFullCRUD() []string {
	return []string{AttrValueCRUDc, AttrValueCRUDr, AttrValueCRUDu, AttrValueCRUDd}
}

type Attr struct {
	Key   AttrKeyType
	Do    AttrDoType
	Value []string
}

func ParseAttr(v string) (*Attr, error) {
	keyVal := strings.Split(v, "=")
	if len(keyVal) != 2 {
		return nil, nil
	}

	attrKey := AttrKeyType(strings.TrimSpace(keyVal[0]))
	attrDo := AttrDoType("")
	attrVal := make([]string, 0)

	doVals := strings.Split(keyVal[1], ":")

	switch len(doVals) {
	case 1:
		if s := strings.TrimSpace(doVals[0]); len(s) > 0 {
			attrVal = append(attrVal, s)
		}
	case 2:
		if s := strings.TrimSpace(doVals[0]); len(s) > 0 {
			attrDo = AttrDoType(s)
		}
		for _, s := range strings.Split(doVals[1], ",") {
			if s = strings.TrimSpace(s); len(s) > 0 {
				attrVal = append(attrVal, s)
			}
		}
	default:
		return nil, fmt.Errorf("invalid attr: got '%s', want format '<key>=<do>:<value>,<value>,...'", v)
	}

	if len(attrDo) == 0 && len(attrVal) == 0 {
		return nil, fmt.Errorf("invalid attr: got '%s', 'key' mast have 'do' or one 'value'", v)
	}

	switch attrKey {
	case AttrKeyTableName:
		if len(attrVal) != 1 {
			return nil, fmt.Errorf("invalid table name: '%s'", v)
		}
		return &Attr{Key: AttrKeyTableName, Value: attrVal}, nil

	case AttrKeyCRUD:
		result := make([]string, 0, 4)
		for _, cv := range GetFullCRUD() {
			for _, dv := range attrVal {
				if strings.Contains(string(dv), string(cv)) {
					result = append(result, cv)
				}
			}
		}
		if len(result) == 0 {
			result = GetFullCRUD()
		}
		return &Attr{Key: AttrKeyCRUD, Value: result}, nil

	case AttrKeyFieldCol:
		if len(attrVal) != 1 {
			return nil, fmt.Errorf("invalid column name: '%s'", v)
		}
		return &Attr{Key: AttrKeyFieldCol, Value: attrVal}, nil

	case AttrKeyFieldLen:
		if len(attrVal) != 1 {
			return nil, fmt.Errorf("invalid column len: '%s'", v)
		}
		return &Attr{Key: AttrKeyFieldLen, Value: attrVal}, nil

	case AttrKeyFieldAuto:
		return &Attr{Key: AttrKeyFieldAuto, Do: attrDo, Value: attrVal}, nil

	case AttrKeyIndex:
		if len(attrDo) == 0 && len(attrVal) == 1 {
			attrDo = AttrDoType(attrVal[0])
			attrVal = []string{}
		}

		switch attrDo {
		case AttrDoIndexPK, AttrDoIndexUniq, AttrDoIndexIdx:
			return &Attr{Key: AttrKeyIndex, Do: attrDo, Value: attrVal}, nil
		case AttrDoIndexFK:
			attrVal = strings.Split(attrVal[0], ".")
			if len(attrVal) != 2 {
				return nil, fmt.Errorf("invalid column foreign key index: '%s'", v)
			}
			return &Attr{Key: AttrKeyIndex, Do: attrDo, Value: attrVal}, nil
		default:
			panic(fmt.Sprintf("unknow index: '%s', must be pk,fk,unq", v))
		}

	default:
		return nil, fmt.Errorf("invalid attribute key: '%s'", v)
	}
}

type Attrs struct {
	data []Attr
}

func NewAttrs() *Attrs {
	return &Attrs{data: make([]Attr, 0)}
}

func (v *Attrs) Set(a Attr) {
	v.data = append(v.data, a)
}

func (v *Attrs) GetByKeyDo(key AttrKeyType, do AttrDoType) (a []Attr, ok bool) {
	for _, datum := range v.data {
		if datum.Key == key && datum.Do == do {
			a = append(a, datum)
			ok = true
		}
	}
	return
}

func (v *Attrs) GetByKey(key AttrKeyType) (a []Attr, ok bool) {
	for _, datum := range v.data {
		if datum.Key == key {
			a = append(a, datum)
			ok = true
		}
	}
	return
}

func (v *Attrs) All() []Attr {
	return v.data
}
