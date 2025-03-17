/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dialects

import "fmt"

type Dialect string

const (
	PGSql  Dialect = "orm:pgsql"
	MySql  Dialect = "orm:mysql"
	SQLite Dialect = "orm:sqlite"
)

func ColComma(dialect Dialect) string {
	switch dialect {
	case PGSql:
		return `"`
	default:
		return "`"
	}
}

func ValComma(dialect Dialect) string {
	return `'`
}

func EscapeCol(dialect Dialect, value string) string {
	comma := ColComma(dialect)
	return comma + value + comma
}

func EscapeVal(dialect Dialect, value string) string {
	comma := ValComma(dialect)
	return comma + value + comma
}

func Vars(dialect Dialect, v int) string {
	switch dialect {
	case PGSql:
		return fmt.Sprintf("$%d", v)
	default:
		return "?"
	}
}
