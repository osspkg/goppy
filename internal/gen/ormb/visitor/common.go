/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package visitor

import (
	"go/ast"

	"go.osspkg.com/goppy/v2/internal/gen/ormb/table"
)

const (
	tag = "gen:orm"
)

func getTypeAndName(v ast.Expr) (fieldType table.FieldType, fieldName string, ok bool) {
	switch vv := v.(type) {
	case *ast.Ident:
		return table.FieldTypeSingle, vv.Name, true

	case *ast.ArrayType:
		fieldType = table.FieldTypeArray
		_, fieldName, ok = getTypeAndName(vv.Elt)
		return

	case *ast.StarExpr:
		fieldType = table.FieldTypeLink
		_, fieldName, ok = getTypeAndName(vv.X)
		return

	case *ast.SelectorExpr:
		fieldType, fieldName, ok = getTypeAndName(vv.X)
		fieldName += "." + vv.Sel.Name
		return

	default:
		return 0, "", false
	}
}
