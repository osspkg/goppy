/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ormb

import (
	"go/ast"
	"go/parser"
	"go/token"

	"go.osspkg.com/console"
	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v3/internal/gen/ormb/common"
	"go.osspkg.com/goppy/v3/internal/gen/ormb/dialects"
	"go.osspkg.com/goppy/v3/internal/gen/ormb/visitor"
	"go.osspkg.com/goppy/v3/internal/global"
	"go.osspkg.com/goppy/v3/orm/dialect"
)

func Command() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("gen-orm", "generate code for orm")
		setter.Flag(func(flagsSetter console.FlagsSetter) {
			flagsSetter.StringVar("dialect", "", "example: orm")
			flagsSetter.StringVar("db-read", "slave", "example: slave")
			flagsSetter.StringVar("db-write", "master", "example: master")
			flagsSetter.StringVar("sql-dir", "", "dir for store sql files")
			flagsSetter.StringVar("model", "Repo", "model name")
			flagsSetter.IntVar("index", 0, "index for sql file as prefix")
		})
		setter.ExecFunc(func(_ []string, _dialect, _dbRead, _dbWrite, _outDir, _modelName string, _index int64) {
			console.Infof("--- GENERATE ---")

			console.ShowDebug(false)

			gen, ok := dialects.Get(dialect.Name(_dialect))
			if !ok {
				console.Fatalf("unknown generate dialect: %s", _dialect)
			}

			currDir := fs.CurrentDir()
			if len(_outDir) == 0 {
				_outDir = currDir
			}

			files, err := fs.SearchFilesByExt(currDir, ".go")
			console.FatalIfErr(err, "search files in %s", currDir)

			for _, filePath := range files {
				console.Debugf("> PARSE FILE: %s", filePath)

				fset := token.NewFileSet()
				f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
				console.FatalIfErr(err, "parse go file: %s", filePath)

				if ast.IsGenerated(f) {
					continue
				}

				vv := &visitor.Visitor{
					Imports:  syncing.NewMap[string, string](1),
					FilePath: filePath[:len(filePath)-3],
				}

				ast.Walk(vv, f)

				cc := common.Config{
					DBRead: _dbRead, DBWrite: _dbWrite,
					CurrDir: currDir, SQLDir: _outDir,
					FileIndex: _index,
					ModelName: _modelName,
				}

				console.FatalIfErr(GenerateSQL(cc, vv, gen), "generate orm sql")
				console.FatalIfErr(GenerateCode(cc, vv, gen), "generate orm code")
			}

			global.ExecPack(true, "gofmt -w -s .", "goimports -l -w .")
		})
	})
}
