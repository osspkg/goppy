/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dialect_pgsql

import (
	"fmt"
	"strings"

	"go.osspkg.com/goppy/v2/internal/gen/ormb/table"
)

type (
	SQL struct {
		E *Escape
	}

	Result struct {
		S, Q, I []string
	}
)

func (q SQL) Build(t *table.Table) (seq, query, index []string, err error) {
	result := &Result{}

	result.Q = append(result.Q, "CREATE TABLE IF NOT EXISTS "+q.E.Cols(t.TableName)+"\n(")
	decr := len(t.Fields)
	for _, field := range t.Fields {
		if err = q.field(t.TableName, field, result); err != nil {
			return
		}
		decr -= 1
		if decr > 0 {
			result.Q[len(result.Q)-1] += ","
		}
	}
	result.Q = append(result.Q, ");")

	if err = q.index(t, result); err != nil {
		return
	}

	return result.S, result.Q, result.I, nil
}

func (q SQL) index(t *table.Table, res *Result) error { //nolint:unparam
	attrs, ok := t.Attrs().GetByKeyDo(table.AttrKeyIndex, table.AttrDoIndexIdx)
	if ok {
		for _, attr := range attrs {
			res.I = append(res.I, "CREATE INDEX "+
				q.E.Cols(t.TableName+"__"+strings.Join(attr.Value, "_")+"__idx")+
				" ON "+q.E.Cols(t.TableName)+" USING btree ("+q.E.Cols(attr.Value...)+")")
		}
	}

	attrs, ok = t.Attrs().GetByKeyDo(table.AttrKeyIndex, table.AttrDoIndexUniq)
	if ok {
		for _, attr := range attrs {
			res.I = append(res.I, "ALTER TABLE "+q.E.Cols(t.TableName)+
				" ADD CONSTRAINT "+q.E.Cols(t.TableName+"__"+strings.Join(attr.Value, "_")+"__unq")+
				" UNIQUE ("+q.E.Cols(attr.Value...)+")")
		}
	}

	return nil
}

func (q SQL) field(t string, f table.TField, res *Result) error {
	var (
		S, Q, C, I []string
	)

	defer func() {
		if len(S) > 0 {
			res.S = append(res.S, strings.Join(S, " "))
		}
		if len(Q) > 0 {
			QR := "\t" + strings.Join(Q, " ")
			if len(C) > 0 {
				QR += ",\n\t" + strings.Join(C, " ")
			}
			res.Q = append(res.Q, QR)
		}
		if len(I) > 0 {
			res.I = append(res.I, strings.Join(I, " "))
		}
	}()

	attrsCol, ok := f.Attrs().GetByKey(table.AttrKeyFieldCol)
	if !ok {
		return fmt.Errorf("column for field %s not found", f.Name())
	}
	col := attrsCol[0].Value[0]
	Q = append(Q, q.E.Cols(col))

	_, isPK := f.Attrs().GetByKeyDo(table.AttrKeyIndex, table.AttrDoIndexPK)
	_, isUNQ := f.Attrs().GetByKeyDo(table.AttrKeyIndex, table.AttrDoIndexUniq)
	_, isIDX := f.Attrs().GetByKeyDo(table.AttrKeyIndex, table.AttrDoIndexIdx)
	attrsFK, isFK := f.Attrs().GetByKeyDo(table.AttrKeyIndex, table.AttrDoIndexFK)
	attrsLen, isLEN := f.Attrs().GetByKey(table.AttrKeyFieldLen)
	noArray := false

	switch f.(type) {
	case table.BigInt:
		Q = append(Q, "BIGINT")
		if isPK {
			noArray = true
			Q = append(Q, "DEFAULT",
				"nextval("+q.E.Vals(t+"__"+col+"__seq")+")")
			S = append(S, "CREATE SEQUENCE IF NOT EXISTS",
				q.E.Cols(t+"__"+col+"__seq"),
				"INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1")
		}
	case table.Int:
		Q = append(Q, "INTEGER")
	case table.SmallInt:
		Q = append(Q, "SMALLINT")
	case table.Chars:
		if isLEN {
			Q = append(Q, "VARCHAR(", attrsLen[0].Value[0], ")")
		} else {
			Q = append(Q, "TEXT")
		}
	case table.UUID:
		Q = append(Q, "UUID")
	case table.Time:
		Q = append(Q, "TIMESTAMPTZ")
	case table.Bool:
		Q = append(Q, "BOOLEAN")
	case table.Real:
		Q = append(Q, "REAL")
	case table.JSONB:
		noArray = true
		Q = append(Q, "JSONB")
	default:
		return fmt.Errorf("unsupported column type: %T", f)
	}

	switch f.Type() {
	case table.FieldTypeSingle:
		Q = append(Q, "NOT NULL")
	case table.FieldTypeArray:
		if !noArray {
			Q = append(Q[:len(Q)-1], Q[len(Q)-1]+"[]")
		}
		Q = append(Q, "NOT NULL")
	case table.FieldTypeLink:
		Q = append(Q, "NULL")
	}

	switch {
	case isPK:
		C = append(C, "CONSTRAINT", q.E.Cols(t+"__"+col+"__pk"),
			"PRIMARY KEY", "(", q.E.Cols(col), ")")
	case isFK:
		C = append(C, "CONSTRAINT", q.E.Cols(t+"__"+col+"__fk"),
			"FOREIGN KEY", "(", q.E.Cols(col), ")",
			"REFERENCES", q.E.Cols(attrsFK[0].Value[0]), "(", q.E.Cols(attrsFK[0].Value[1]), ")",
			"ON DELETE CASCADE NOT DEFERRABLE",
		)
	case isUNQ:
		C = append(C, "CONSTRAINT", q.E.Cols(t+"__"+col+"__unq"),
			"UNIQUE", "(", q.E.Cols(col), ")",
		)
	case isIDX:
		I = append(I, "CREATE INDEX",
			q.E.Cols(t+"__"+col+"__idx"),
			"ON", q.E.Cols(t), "USING btree (", q.E.Cols(col), ")")
	}

	return nil
}
