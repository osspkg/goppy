/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package visitor

import (
	"go/ast"
	"strings"

	"go.osspkg.com/console"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v2/internal/gen/ormb/common"
	"go.osspkg.com/goppy/v2/internal/gen/ormb/table"
)

type Visitor struct {
	FilePath string
	PkgName  string
	Imports  *syncing.Map[string, string]
	Tables   []*table.Table
}

func (v *Visitor) parseComment(comment string, attrs *table.Attrs) {
	comment = strings.TrimPrefix(comment, "//")
	comment = strings.TrimSpace(comment)

	console.Debugf("-- parse comment: %s", comment)

	for _, s := range strings.Fields(comment) {
		a, err := table.ParseAttr(s)
		console.FatalIfErr(err, "parse comment: %s", s)

		if a != nil {
			attrs.Set(*a)
		}
	}
}

func (v *Visitor) parseField(field *ast.Field) (result []table.TField) {
	for _, name := range field.Names {
		console.Debugf("Parse field: %s", name)

		if !name.IsExported() || field.Comment == nil {
			console.Debugf("---- skip")
			continue
		}

		fieldType, goType, ok := getTypeAndName(field.Type)
		if !ok {
			console.Debugf("-- type & name not found: %s", field.Type)
			continue
		}

		fieldItem := table.CreateFieldType(fieldType, name.String(), goType)
		console.Debugf("-- create field: %s=%s", name.String(), goType)

		for _, comment := range field.Comment.List {
			v.parseComment(comment.Text, fieldItem.Attrs())
		}

		if _, ok = fieldItem.Attrs().GetByKey(table.AttrKeyFieldCol); !ok {
			console.Debugf("---- col not found")
			continue
		}

		result = append(result, fieldItem)
		console.Debugf("---- added: %s", fieldItem.String())
	}

	return
}

func (v *Visitor) astFile(node *ast.File) ast.Visitor {
	v.PkgName = node.Name.String()

	console.Debugf("Parsed PkgName: %s", v.PkgName)

	return v
}

func (v *Visitor) astImportSpec(node *ast.ImportSpec) ast.Visitor {
	path := strings.Trim(node.Path.Value, `"`)
	name := common.SplitLast(path, "/")

	if node.Name != nil {
		name = node.Name.String()
	}

	console.Debugf("Import: name='%s' path='%s'", name, path)

	v.Imports.Set(name, path)

	return v
}

func (v *Visitor) astTypeSpec(node *ast.TypeSpec) {
	structNode, ok := node.Type.(*ast.StructType)
	if !ok || node.Doc == nil || len(structNode.Fields.List) == 0 || node.Name == nil {
		return
	}

	model := table.CreateTableType()
	model.ModelName = node.Name.String()

	console.Debugf("* Parse model: %s", model.ModelName)

	for _, doc := range node.Doc.List {
		i := strings.Index(doc.Text, tag)
		if i < 0 {
			continue
		}
		v.parseComment(doc.Text[i+len(tag):], model.Attrs())
	}

	tableName, ok := model.Attrs().GetByKey(table.AttrKeyTableName)
	if !ok {
		return
	}

	model.TableName = string(tableName[0].Value[0])

	for _, field := range structNode.Fields.List {
		model.Fields = append(model.Fields, v.parseField(field)...)
	}

	if len(model.Fields) == 0 {
		return
	}

	v.Tables = append(v.Tables, model)

	return
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	switch nodeType := node.(type) {
	case *ast.File:
		return v.astFile(nodeType)

	case *ast.ImportSpec:
		return v.astImportSpec(nodeType)

	case *ast.GenDecl:
		for _, spec := range nodeType.Specs {
			switch specType := spec.(type) {
			case *ast.TypeSpec:
				if nodeType.Doc != nil {
					specType.Doc = nodeType.Doc
				}
				v.astTypeSpec(specType)
			}
		}
		return v

	default:
		return nil
	}
}
