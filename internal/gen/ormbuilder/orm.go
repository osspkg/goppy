/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ormbuilder

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/code"
	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/common"
	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/sql"
	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/visitor"
)

func New(c common.Config) error {
	files, e := fs.SearchFilesByExt(c.Dir, ".go")
	if e != nil {
		return e
	}

	fset := token.NewFileSet()

	for _, filePath := range files {
		f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		if ast.IsGenerated(f) {
			continue
		}

		v := &visitor.Visitor{
			Imports:  syncing.NewMap[string, string](1),
			FilePath: filePath[:len(filePath)-3],
		}

		ast.Walk(v, f)

		fmt.Println("IMPORT:", v.Imports.Keys())
		for _, model := range v.Models {
			fmt.Println("MODEL:", model.Name, model.Table, model.Attr.All())
			for _, field := range model.Fields {
				fmt.Println(" --- ", field.Name(), field.Col(), field.RawType(), field.Type(), field.Attr().All())
			}
		}

		if err = sql.Generate(c, v); err != nil {
			return err
		}

		if err = code.Generate(c, v); err != nil {
			return err
		}
	}

	return nil
}
