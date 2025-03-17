/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package visitor

import (
	"go/ast"

	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/fields"
)

func GetTypeAndName(v ast.Expr) (ft fields.FieldType, name string, ok bool) {
	switch v := v.(type) {
	case *ast.Ident:
		return fields.FieldTypeSingle, v.Name, true
	case *ast.ArrayType:
		ft = fields.FieldTypeArray
		_, name, ok = GetTypeAndName(v.Elt)
		return
	case *ast.StarExpr:
		ft = fields.FieldTypeLink
		_, name, ok = GetTypeAndName(v.X)
		return
	case *ast.SelectorExpr:
		ft, name, ok = GetTypeAndName(v.X)
		name += "." + v.Sel.Name
		return
	default:
		return 0, "", false
	}
}
