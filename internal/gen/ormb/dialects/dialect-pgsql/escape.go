/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dialect_pgsql

import (
	"fmt"
	"strings"
)

const (
	colComma = `"`
	valComma = `'`
)

type Escape struct{}

func (Escape) ColComma() string { return colComma }
func (Escape) ValComma() string { return valComma }

func (e Escape) Cols(values ...string) string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, colComma+strings.TrimSpace(value)+colComma)
	}
	return strings.Join(result, ", ")
}

func (e Escape) Vals(values ...string) string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, valComma+strings.TrimSpace(value)+valComma)
	}
	return strings.Join(result, ", ")
}

func (e Escape) Vars(ns ...int) string {
	result := make([]string, 0, len(ns))
	for _, n := range ns {
		result = append(result, fmt.Sprintf("$%d", n))
	}
	return strings.Join(result, ", ")
}

func (e Escape) VarsRangeStr(from, to int) string {
	return strings.Join(e.VarsRange(from, to), ", ")
}

func (e Escape) VarsRange(from, to int) []string {
	result := make([]string, 0, to-from)
	for i := from; i <= to; i++ {
		result = append(result, fmt.Sprintf("$%d", i))
	}
	return result
}
