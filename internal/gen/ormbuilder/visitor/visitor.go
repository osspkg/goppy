/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package visitor

import (
	"go/ast"
	"strings"

	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/common"
	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/fields"
)

type Visitor struct {
	PkgName  string
	Imports  *syncing.Map[string, string]
	FilePath string
	Models   []*Model
}

func (v *Visitor) parseComment(c string, attrs *fields.Attrs) {
	for _, s := range strings.Fields(c) {
		a, ok := fields.ParseAttr(s)
		if !ok {
			continue
		}
		attrs.Set(*a)
	}
}

func (v *Visitor) parseField(field *ast.Field) (result []fields.TField) {
	for _, name := range field.Names {
		if !name.IsExported() {
			continue
		}

		ft, val, ok := GetTypeAndName(field.Type)
		if !ok {
			continue
		}

		if field.Comment == nil {
			continue
		}

		attrs := fields.NewAttrs()
		for _, comment := range field.Comment.List {
			v.parseComment(comment.Text, attrs)
		}

		colName, _ := attrs.FirstValue(fields.AttrFieldCol)

		fieldItem := fields.Create(val, ft, colName, name.String())

		for _, item := range attrs.All() {
			switch item.Type {
			case fields.AttrFieldCol:
				continue

			case fields.AttrIndexPK, fields.AttrIndexUniq:
				item.Value = []string{colName}
			}

			fieldItem.Attr().Set(item)
		}

		result = append(result, fieldItem)
	}

	return
}

func (v *Visitor) astFile(node *ast.File) ast.Visitor {
	v.PkgName = node.Name.String()

	return v
}

func (v *Visitor) astImportSpec(node *ast.ImportSpec) ast.Visitor {
	path := strings.Trim(node.Path.Value, `"`)
	name := common.SplitLast(path, "/")

	if node.Name != nil {
		name = node.Name.String()
	}

	v.Imports.Set(name, path)

	return v
}

func (v *Visitor) astTypeSpec(node *ast.TypeSpec) {
	structNode, ok := node.Type.(*ast.StructType)
	if !ok || node.Doc == nil || len(structNode.Fields.List) == 0 || node.Name == nil {
		return
	}

	model := NewModel()
	model.Name = fields.ModelNameType(node.Name.String())

	for _, doc := range node.Doc.List {
		if !strings.Contains(doc.Text, tag) {
			continue
		}
		v.parseComment(doc.Text, model.Attr)
	}

	if table, ok := model.Attr.FirstValue(fields.AttrTableName); ok {
		model.Table = fields.TableType(table)
	}

	for _, field := range structNode.Fields.List {
		model.Fields = append(model.Fields, v.parseField(field)...)
	}

	v.Models = append(v.Models, model)

	return
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.File:
		return v.astFile(node)

	case *ast.ImportSpec:
		return v.astImportSpec(node)

	case *ast.GenDecl:
		for _, spec := range node.Specs {
			switch spec := spec.(type) {
			case *ast.TypeSpec:
				if node.Doc != nil {
					spec.Doc = node.Doc
				}
				v.astTypeSpec(spec)
			}
		}
		return v

	default:
		return nil
	}
}
