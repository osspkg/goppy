/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ormb

import (
	"fmt"
	"os"
	"strings"

	"go.osspkg.com/console"
	"go.osspkg.com/ioutils/data"

	"go.osspkg.com/goppy/v3/internal/gen/ormb/common"
	"go.osspkg.com/goppy/v3/internal/gen/ormb/dialects"
	"go.osspkg.com/goppy/v3/internal/gen/ormb/visitor"
)

func GenerateSQL(cc common.Config, vv *visitor.Visitor, g *dialects.Gen) error {
	i := cc.FileIndex
	w := data.NewBuffer(1024)

	for _, tab := range vv.Tables {
		w.Reset()

		console.Debugf(">> Gen Table:\n%s\n", tab.String())

		seq, query, index, err := g.SQL.Build(tab)
		if err != nil {
			return err
		}

		if len(seq) > 0 {
			common.Write(w, "-- SEQUENCE\n")
			common.Write(w, strings.Join(seq, ";\n"))
			common.Write(w, ";\n\n")
		}

		if len(query) > 0 {
			common.Write(w, "-- TABLE\n")
			common.Write(w, strings.Join(query, "\n"))
			common.Write(w, "\n\n")
		}

		if len(index) > 0 {
			common.Write(w, "-- INDEX\n")
			common.Write(w, strings.Join(index, ";\n"))
			common.Write(w, ";\n\n")
		}

		if err = os.WriteFile(
			fmt.Sprintf("%s/%06d_%s_table.sql", cc.SQLDir, i, tab.TableName),
			w.Bytes(), 0755,
		); err != nil {
			return err
		}

		i++
	}

	return nil
}
