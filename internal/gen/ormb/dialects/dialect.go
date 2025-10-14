/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dialects

import (
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v2/internal/gen/ormb/common"
	dialectpgsql "go.osspkg.com/goppy/v2/internal/gen/ormb/dialects/dialect-pgsql"
	"go.osspkg.com/goppy/v2/internal/gen/ormb/table"
	"go.osspkg.com/goppy/v2/orm/clients/pgsql"
	"go.osspkg.com/goppy/v2/orm/dialect"
)

var (
	escapeMap = syncing.NewMap[dialect.Name, *Gen](3)
)

type (
	TEscape interface {
		ColComma() string
		ValComma() string
		Cols(values ...string) string
		Vals(values ...string) string
		Vars(ns ...int) string
		VarsRangeStr(from, to int) string
		VarsRange(from, to int) []string
	}

	TSql interface {
		Build(t *table.Table) (seq, query, index []string, err error)
	}

	TCode interface {
		Build(t *table.Table, ci common.CodeInfo) []byte
	}

	Gen struct {
		Escape TEscape
		SQL    TSql
		Code   TCode
	}
)

func init() {
	escapeMap.Set(pgsql.Name, func() *Gen {
		e := &dialectpgsql.Escape{}
		return &Gen{Escape: e, SQL: &dialectpgsql.SQL{E: e}, Code: &dialectpgsql.Code{E: e}}
	}())
}

func Get(name dialect.Name) (*Gen, bool) {
	return escapeMap.Get(name)
}
