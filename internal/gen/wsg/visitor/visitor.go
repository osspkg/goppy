/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package visitor

import (
	"errors"
	"fmt"
	"go/ast"
	"io"
	"path/filepath"
	"strings"

	"go.osspkg.com/bb"
	"go.osspkg.com/do"
	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v3/console"
	"go.osspkg.com/goppy/v3/internal/gen/ormb/common"
	"go.osspkg.com/goppy/v3/internal/gen/wsg/types"
)

const tag = "@wsg"

type Visitor struct {
	FilePath string
	PkgName  string
	PkgPath  string
	GoMod    string
	Imports  *syncing.Map[string, string]
	Objects  []types.Object
}

func (v *Visitor) ToFile() types.File {
	return types.File{
		FilePath: v.FilePath,
		PkgName:  v.PkgName,
		PkgPath:  v.PkgPath,
		GoMod:    v.GoMod,
		Imports:  v.Imports,
		Objects:  v.Objects,
	}
}

func (v *Visitor) Debug() {
	fmt.Println("=============================================================")
	fmt.Println("FilePath:", strings.TrimPrefix(v.FilePath, fs.CurrentDir()))
	fmt.Println("PkgName:", v.PkgName, "PkgPath:", v.PkgPath)
	fmt.Println("Import:")
	for alias, path := range v.Imports.Yield() {
		fmt.Println("  ", alias, path)
	}
	for _, object := range v.Objects {
		fmt.Println("Interface:", object.Name)
		fmt.Println("  tags:")
		for key, value := range object.Tags {
			fmt.Println("    ", key, ":", value)
		}

		for _, method := range object.Methods {
			fmt.Println("  method:", method.Name)

			fmt.Println("    tags:")
			for key, value := range method.Tags {
				fmt.Println("      ", key, ":", value)
			}

			fmt.Println("    in:")
			for _, value := range method.InParams {
				fmt.Println("      ",
					"name:", value.Name, ", type:", value.Type,
					", pkg:", value.Pkg, ", omit:", value.Omitempty)
			}

			fmt.Println("    out:")
			for _, value := range method.OutParams {
				fmt.Println("      ",
					"name:", value.Name, ", type:", value.Type,
					", pkg:", value.Pkg, ", omit:", value.Omitempty)
			}
		}
	}
	fmt.Println("=============================================================")
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

func (v *Visitor) astFile(node *ast.File) ast.Visitor {
	v.PkgName = node.Name.String()
	v.PkgPath = v.GoMod + filepath.Dir(v.FilePath)

	console.Debugf("Parsed PkgName: %s -> %s", v.PkgName, v.PkgPath)

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

func (v *Visitor) parseDoc(comment *ast.CommentGroup, tags *types.Tags) {
	if comment == nil {
		return
	}
	for _, doc := range comment.List {
		i := strings.Index(doc.Text, tag)
		if i < 0 {
			continue
		}
		v.parseComment(doc.Text[i+len(tag):], tags)
	}
}

func (v *Visitor) parseComment(comment string, tags *types.Tags) {
	comment = strings.TrimPrefix(comment, "//")
	comment = strings.TrimSpace(comment)

	console.Debugf("-- parse comment: %s", comment)

	buf := bb.FromBytes([]byte(comment))
	_, err := buf.Seek(0, io.SeekStart)
	console.FatalIfErr(err, "parse comment")

	for {
		key, err := buf.ReadString('=')
		if errors.Is(err, io.EOF) {
			break
		}
		console.FatalIfErr(err, "parse comment")
		key = strings.Trim(strings.TrimSpace(key), "=")

		r, _, err := buf.ReadRune()
		if errors.Is(err, io.EOF) {
			break
		}
		console.FatalIfErr(err, "parse comment")

		var value string
		switch r {
		case '"':
			value, err = buf.ReadString('"')
			value = strings.Trim(strings.TrimSpace(value), "\"")
		case '\'':
			value, err = buf.ReadString('\'')
			value = strings.Trim(strings.TrimSpace(value), "'")
		case '`':
			value, err = buf.ReadString('`')
			value = strings.Trim(strings.TrimSpace(value), "`")
		default:
			console.FatalIfErr(buf.UnreadRune(), "parse comment")
			value, err = buf.ReadString(' ')
		}

		if errors.Is(err, io.EOF) {
			break
		}
		console.FatalIfErr(err, "parse comment")

		(*tags)[key] = append((*tags)[key], do.Convert[string, string](
			strings.Split(value, ","),
			func(value string, index int) string {
				return strings.TrimSpace(value)
			},
		)...)
	}
}

func (v *Visitor) parseMethods(fields *ast.FieldList) (result []types.Method) {
	for _, field := range fields.List {
		if !field.Names[0].IsExported() {
			continue
		}

		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		console.Debugf("** Parse method: %v", field.Names[0].String())

		method := types.Method{
			Name:      field.Names[0].String(),
			Tags:      make(types.Tags, 10),
			InParams:  nil,
			OutParams: nil,
		}

		v.parseDoc(field.Doc, &method.Tags)
		//v.parseDoc(field.Comment, &method.Tags)

		if funcType.Params != nil {
			for _, param := range funcType.Params.List {
				method.InParams = append(method.InParams, getParam(param))
			}
		}

		if funcType.Results != nil {
			for _, param := range funcType.Results.List {
				method.OutParams = append(method.OutParams, getParam(param))
			}
		}

		result = append(result, method)

	}

	return
}

func (v *Visitor) astTypeSpec(node *ast.TypeSpec) {
	faceNode, ok := node.Type.(*ast.InterfaceType)
	if !ok {
		return
	}

	obj := types.Object{
		Pkg:   v.PkgPath,
		Alias: v.PkgName,
		Name:  node.Name.String(),
		Tags:  make(types.Tags, 10),
	}

	console.Debugf("* Parse interface: %s", obj.Name)

	v.parseDoc(node.Doc, &obj.Tags)
	//v.parseDoc(node.Comment, &obj.Tags)

	obj.Methods = append(obj.Methods, v.parseMethods(faceNode.Methods)...)

	v.Objects = append(v.Objects, obj)

	return
}

func getParam(param *ast.Field) types.Param {
	paramType := getTypeName(param.Type)
	paramPkg := func() string {
		list := strings.Split(paramType, ".")
		switch len(list) {
		case 1:
			return ""
		case 2:
			paramType = list[1]
			return strings.Trim(list[0], "*[].")
		default:
			console.Fatalf("invalid type: %s", paramType)
			return ""
		}
	}()

	console.Debugf("---- parse arg: Name: %s, Type: %s, Pkg: %s",
		param.Names[0].String(), paramType, paramPkg)

	return types.Param{
		Name: param.Names[0].String(),
		Type: paramType,
		Pkg:  paramPkg,
		Omitempty: strings.HasPrefix(paramType, "*") ||
			strings.HasPrefix(paramType, "[]"),
	}
}

/*
TODO: change to
import "go/printer"
import "strings"
import "bytes"

func exprToString(fset *token.FileSet, expr ast.Expr) string {
    var buf bytes.Buffer
    printer.Fprint(&buf, fset, expr)
    return buf.String()
}
*/

func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", getTypeName(t.X), t.Sel.Name)
	case *ast.StarExpr:
		return "*" + getTypeName(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeName(t.Elt)
	default:
		return ""
	}
}
