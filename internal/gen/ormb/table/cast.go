/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package table

type BigInt struct{ base }
type Int struct{ base }
type SmallInt struct{ base }
type Chars struct{ base }
type UUID struct{ base }
type Time struct{ base }
type Bool struct{ base }
type Real struct{ base }
type JSONB struct{ base }

func CreateFieldType(fieldType FieldType, fieldName string, goType string) TField {
	b := base{
		name:      fieldName,
		fieldType: fieldType,
		goType:    goType,
		attr:      NewAttrs(),
	}

	switch goType {
	case "int32", "uint32":
		return Int{b}
	case "int16", "uint16", "int8", "uint8":
		return SmallInt{b}
	case "int64", "int", "uint64", "uint":
		return BigInt{b}
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
