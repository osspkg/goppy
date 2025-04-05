/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package gen

import (
	"strings"

	"go.osspkg.com/console"
	"go.osspkg.com/ioutils/fs"

	"go.osspkg.com/goppy/v2/internal/global"

	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/common"
	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/dialects"

	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder"
)

func Command() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("gen", "generate code")
		setter.Flag(func(flagsSetter console.FlagsSetter) {
			flagsSetter.StringVar("type", "", "example: orm")
			flagsSetter.StringVar("db-read", "slave", "example: slave")
			flagsSetter.StringVar("db-write", "master", "example: master")
			flagsSetter.StringVar("sql-dir", "", "dir for store sql files")
			flagsSetter.IntVar("index", 0, "index for sql file as prefix")
		})
		setter.ExecFunc(func(_ []string, types, dbr, dbw, outDir string, index int64) {
			console.Infof("--- GENERATE ---")

			dir := fs.CurrentDir()
			list := strings.Split(types, ",")
			if len(outDir) == 0 {
				outDir = dir
			}

			for _, s := range list {
				dialect := dialects.Dialect(strings.TrimSpace(strings.ToLower(s)))

				switch dialect {
				case dialects.PGSql:
					console.FatalIfErr(
						ormbuilder.New(common.Config{
							Dialect: dialect,
							DBRead:  dbr, DBWrite: dbw,
							Dir: dir, SQLDir: outDir,
							FileIndex: index,
						}),
						"generate orm preset")

				default:
					console.Fatalf("unknown generate type: %s", s)
				}

				global.ExecPack("gofmt -w -s .", "goimports -l -w .")
			}

		})
	})
}
