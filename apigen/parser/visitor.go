/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"path/filepath"
	"strings"

	"go.osspkg.com/bb"
	"go.osspkg.com/do"
	"go.osspkg.com/errors"
	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v3/apigen/types"
	"go.osspkg.com/goppy/v3/apigen/util"
	"go.osspkg.com/goppy/v3/internal/global"
)

type (
	visitor struct {
		TAG string

		FilePath string
		PkgName  string
		PkgPath  string
		GoMod    string
		Imports  *syncing.Map[string, string]
		Objects  []types.Face
	}

	Parser interface {
		ToFile() types.File
		DumpStdout()
	}
)

func New(tag, filePath string) (Parser, error) {
	gomod, root, err := global.DetectGoMod(filePath)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	if ast.IsGenerated(f) {
		return nil, ErrIsGenerated
	}

	vis := &visitor{
		TAG:      tag,
		Imports:  syncing.NewMap[string, string](10),
		FilePath: strings.TrimPrefix(filePath, root),
		GoMod:    gomod,
	}

	ast.Walk(vis, f)

	return vis, nil
}

func (v *visitor) ToFile() types.File {
	return types.File{
		FilePath: v.FilePath,
		PkgName:  v.PkgName,
		PkgPath:  v.PkgPath,
		GoMod:    v.GoMod,
		Imports:  v.Imports,
		Faces:    v.Objects,
	}
}

func (v *visitor) DumpStdout() {
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
					", pkg:", value.Pkg, ", omit:", value.Omitempty,
					", ptr:", value.Ptr, ", slice:", value.Slice)
			}

			fmt.Println("    out:")
			for _, value := range method.OutParams {
				fmt.Println("      ",
					"name:", value.Name, ", type:", value.Type,
					", pkg:", value.Pkg, ", omit:", value.Omitempty,
					", ptr:", value.Ptr, ", slice:", value.Slice)
			}
		}
	}
	fmt.Println("=============================================================")
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
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

func (v *visitor) astFile(node *ast.File) ast.Visitor {
	v.PkgName = node.Name.String()
	v.PkgPath = v.GoMod + filepath.Dir(v.FilePath)

	v.Imports.Set(v.PkgName, v.PkgPath)

	return v
}

func (v *visitor) astImportSpec(node *ast.ImportSpec) ast.Visitor {
	path := strings.Trim(node.Path.Value, `"`)
	name := util.SplitLast(path, "/")

	if node.Name != nil {
		name = node.Name.String()
	}

	v.Imports.Set(name, path)

	return v
}

func (v *visitor) parseDoc(comment *ast.CommentGroup, tags *types.Tags) {
	if comment == nil {
		return
	}
	for _, doc := range comment.List {
		i := strings.Index(doc.Text, v.TAG)
		if i < 0 {
			continue
		}
		v.parseComment(doc.Text[i+len(v.TAG):], tags)
	}
}

func (v *visitor) parseComment(comment string, tags *types.Tags) {
	comment = strings.TrimPrefix(comment, "//")
	comment = strings.TrimSpace(comment)

	buf := bb.FromBytes([]byte(comment))
	_, err := buf.Seek(0, io.SeekStart)
	util.PanicIfError(err, "parse comment")

	for {
		key, err := buf.ReadString('=')
		if errors.Is(err, io.EOF) {
			break
		}
		util.PanicIfError(err, "parse comment")
		key = strings.Trim(strings.TrimSpace(key), "=")

		r, _, err := buf.ReadRune()
		if errors.Is(err, io.EOF) {
			break
		}
		util.PanicIfError(err, "parse comment")

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
			util.PanicIfError(buf.UnreadRune(), "parse comment")
			value, err = buf.ReadString(' ')
		}

		if errors.Is(err, io.EOF) {
			break
		}
		util.PanicIfError(err, "parse comment")

		(*tags)[key] = append((*tags)[key], do.Convert[string, string](
			strings.Split(value, ","),
			func(value string, index int) string {
				return strings.TrimSpace(value)
			},
		)...)
	}
}

func (v *visitor) parseMethods(fields *ast.FieldList) (result []types.Method) {
	for _, field := range fields.List {
		if !field.Names[0].IsExported() {
			continue
		}

		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue
		}

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
				rp := getParam(param)
				if !isBaseType(rp.Type) && len(rp.Pkg) == 0 {
					rp.Pkg = v.PkgName
				}

				method.InParams = append(method.InParams, rp)
			}
		}

		if funcType.Results != nil {
			for _, param := range funcType.Results.List {
				rp := getParam(param)
				if !isBaseType(rp.Type) && len(rp.Pkg) == 0 {
					rp.Pkg = v.PkgName
				}

				method.OutParams = append(method.OutParams, rp)
			}
		}

		result = append(result, method)

	}

	return
}

func (v *visitor) astTypeSpec(node *ast.TypeSpec) {
	faceNode, ok := node.Type.(*ast.InterfaceType)
	if !ok {
		return
	}

	obj := types.Face{
		Pkg:   v.PkgPath,
		Alias: v.PkgName,
		Name:  node.Name.String(),
		Tags:  make(types.Tags, 10),
	}

	v.parseDoc(node.Doc, &obj.Tags)
	//v.parseDoc(node.Comment, &obj.Tags)

	obj.Methods = append(obj.Methods, v.parseMethods(faceNode.Methods)...)

	v.Objects = append(v.Objects, obj)

	return
}

func getParam(param *ast.Field) (p types.Param) {
	p.Name = param.Names[0].String()
	p.Type = getTypeName(param.Type)
	p.Slice = strings.HasPrefix(p.Type, "[]")
	p.Ptr = strings.HasPrefix(p.Type, "*")
	p.Omitempty = p.Ptr || p.Slice

	list := strings.Split(p.Type, ".")
	switch len(list) {
	case 1:
		p.Type = strings.Trim(list[0], "*[].")
	case 2:
		p.Type = list[1]
		p.Pkg = strings.Trim(list[0], "*[].")
	default:
		util.PanicIfError(fmt.Errorf("invalid type: %s", p.Type), "get param")
	}

	return
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

func isBaseType(arg string) bool {
	switch arg {
	case "bool", "string", "byte", "struct", "struct{}",
		"any", "interface", "interface{}",
		"complex64", "complex128",
		"error", "uintptr",
		"float32", "float64",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"map":
		return true
	default:
		return false
	}
}
