/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package fields

type Number struct{ base }

type Chars struct{ base }

type UUID struct{ base }

type Time struct{ base }

type Bool struct{ base }

type Real struct{ base }

type JSONB struct{ base }

func Create(rt string, t FieldType, c, n string) TField {

	b := base{
		name:      n,
		col:       c,
		fieldType: t,
		rawType:   rt,
		attr:      NewAttrs(),
	}

	switch rt {
	case "int64", "int32", "int", "uint64", "uint32", "uint":
		return Number{b}
	case "byte", "string":
		return Chars{b}
	case "uuid.UUID":
		return UUID{b}
	case "time.Time":
		return Time{b}
	case "bool":
		return Bool{b}
	case "float64", "float32":
		return Real{b}
	default:
		return JSONB{b}
	}
}
