/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package sql

import (
	"fmt"
	"os"
	"strings"

	"go.osspkg.com/do"
	"go.osspkg.com/ioutils/data"

	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/common"
	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/dialects"
	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/fields"
	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/visitor"
)

func Generate(c common.Config, v *visitor.Visitor) (err error) {
	index := c.FileIndex
	for _, m := range v.Models {
		if len(m.Table) == 0 {
			continue
		}

		//if attr, ok := m.Attr.FirstValue(fields.AttrAction); ok {
		//	switch attr {
		//	case fields.AttrValueActionRO:
		//		continue
		//	}
		//}

		w := data.NewBuffer(1024)

		var (
			query []string
			seq   []string
		)

		for _, f := range m.Fields {
			if len(f.Col()) == 0 {
				continue
			}

			tmpQuery, tmpSeq := field(c.Dialect, string(m.Table), f)

			if len(tmpQuery) > 0 {
				query = append(query, strings.Join(tmpQuery, " "))
			}
			if len(tmpSeq) > 0 {
				seq = append(seq, strings.Join(tmpSeq, " "))
			}
		}

		if list, ok := m.Attr.Get(fields.AttrIndexUniq); ok {
			for _, attr := range list {
				cols := make([]string, 0, len(attr.Value))
				for _, s := range attr.Value {
					cols = append(cols, dialects.EscapeCol(c.Dialect, s))
				}
				query = append(query,
					"\t CONSTRAINT "+dialects.EscapeCol(c.Dialect,
						fmt.Sprintf("%s__%s__uniq", string(m.Table), strings.Join(attr.Value, "_")))+
						" UNIQUE ("+strings.Join(cols, ",")+")",
				)
			}
		}

		if len(seq) > 0 {
			common.Write(w, "-- SEQUENCE\n")
			common.Write(w, strings.Join(seq, ";\n"))
			common.Write(w, ";\n\n")
		}
		if len(query) > 0 {
			common.Write(w, "-- TABLE\n")
			common.Write(w, "CREATE TABLE IF NOT EXISTS %s (\n", dialects.EscapeCol(c.Dialect, string(m.Table)))
			common.Write(w, strings.Join(query, ",\n"))
			common.Write(w, "\n);\n\n")
		}

		if err = os.WriteFile(
			fmt.Sprintf("%s/%06d_%s_table.sql", c.SQLDir, index, m.Table),
			w.Bytes(), 0755,
		); err != nil {
			return err
		}

		index++
	}
	return nil
}

func field(dialect dialects.Dialect, table string, v fields.TField) (query []string, seq []string) {
	query = append(query, "\t", dialects.EscapeCol(dialect, v.Col()))

	nonArr := false
	_, isPK := v.Attr().FirstValue(fields.AttrIndexPK)
	_, isUNIQ := v.Attr().FirstValue(fields.AttrIndexUniq)
	valFK, isFK := v.Attr().FirstValue(fields.AttrIndexFK)
	valLen, hasLen := v.Attr().FirstValue(fields.AttrFieldLen)

	switch v.(type) {

	case fields.Number:
		switch dialect {
		case dialects.PGSql:
			query = append(query, "BIGINT")
			if isPK {
				query = append(query,
					"DEFAULT",
					"nextval('"+table+"_"+v.Col()+"_seq')",
				)
				seq = append(seq,
					"CREATE SEQUENCE IF NOT EXISTS",
					dialects.EscapeCol(dialect, table+"_"+v.Col()+"_seq"),
					"INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1",
				)
			}
		case dialects.MySql:
		case dialects.SQLite:
		}

	case fields.Chars:
		nonArr = hasLen
		switch dialect {
		case dialects.PGSql:
			query = append(query, do.IfElse(hasLen, "VARCHAR("+valLen+")", "TEXT"))
		case dialects.MySql:
		case dialects.SQLite:
		}

	case fields.UUID:
		switch dialect {
		case dialects.PGSql:
			query = append(query, "UUID")
		case dialects.MySql:
		case dialects.SQLite:
		}

	case fields.Time:
		switch dialect {
		case dialects.PGSql:
			query = append(query, "TIMESTAMPTZ")
		case dialects.MySql:
		case dialects.SQLite:
		}

	case fields.Bool:
		switch dialect {
		case dialects.PGSql:
			query = append(query, "BOOLEAN")
		case dialects.MySql:
		case dialects.SQLite:
		}

	case fields.Real:
		switch dialect {
		case dialects.PGSql:
			query = append(query, "REAL")
		case dialects.MySql:
		case dialects.SQLite:
		}

	case fields.JSONB:
		nonArr = true
		switch dialect {
		case dialects.PGSql:
			query = append(query, "JSONB")
		case dialects.MySql:
		case dialects.SQLite:
		}

	default:
		panic("unknown sql type")
	}

	switch v.Type() {
	case fields.FieldTypeSingle:
		query = append(query, "NOT NULL")

	case fields.FieldTypeArray:
		switch dialect {
		case dialects.PGSql:
			if !nonArr {
				query = append(query[:len(query)-1], query[len(query)-1]+"[]")
			}
		case dialects.MySql:
		case dialects.SQLite:
		}
		query = append(query, "NOT NULL")

	case fields.FieldTypeLink:
		query = append(query, "NULL")
	}

	if isPK {
		query = append(query, ",\n\t",
			"CONSTRAINT", dialects.EscapeCol(dialect, table+"_"+v.Col()+"_pk"),
			"PRIMARY KEY", "("+dialects.EscapeCol(dialect, v.Col())+")",
		)
		return
	}

	if isUNIQ {
		query = append(query, ",\n\t",
			"CONSTRAINT", dialects.EscapeCol(dialect, table+"_"+v.Col()+"_unq"),
			"UNIQUE", "("+dialects.EscapeCol(dialect, v.Col())+")",
		)
	}

	if isFK {
		fk := strings.Split(valFK, ".")
		if len(fk) != 2 {
			return
		}
		query = append(query, ",\n\t",
			"CONSTRAINT", dialects.EscapeCol(dialect, table+"_"+v.Col()+"_fk"),
			"FOREIGN KEY", "("+dialects.EscapeCol(dialect, v.Col())+")",
			"REFERENCES", dialects.EscapeCol(dialect, fk[0]), "("+dialects.EscapeCol(dialect, fk[1])+")",
			"ON DELETE CASCADE NOT DEFERRABLE",
		)
	}

	return
}
