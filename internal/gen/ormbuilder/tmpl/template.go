/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package tmpl

import (
	"io"
	"strings"
	"text/template"

	"go.osspkg.com/goppy/v2/internal/gen/ormbuilder/dialects"
)

type Escaped string
type Repeatable int

func Build(w io.Writer, value string, model any) error {
	tmpl, err := template.
		New("0").
		Funcs(template.FuncMap{
			"esc":     esc,
			"lower":   strings.ToLower,
			"title":   title,
			"queries": queries,
			"fields":  fields,
			"pls":     pls,
			"incr":    incr,
		}).
		Parse(value)

	if err != nil {
		return err
	}

	return tmpl.Execute(w, model)
}

func esc(dialect string, arg string) string {
	return dialects.EscapeCol(dialects.Dialect(dialect), arg)
}

func incr(v, i int) int {
	return v + i
}

func title(arg string) string {
	arg = strings.ReplaceAll(arg, `_`, ` `)
	arg = strings.ReplaceAll(arg, `-`, ` `)
	args := strings.Fields(arg)
	result := make([]string, 0, len(args))
	for _, s := range args {
		result = append(result, strings.ToUpper(s[:1])+strings.ToLower(s[1:]))
	}
	return strings.Join(result, "")
}

func queries(dialect string, args ...any) string {
	var sb strings.Builder

	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			sb.WriteString(v)

		case []string:
			for i, s := range v {
				sb.WriteString(s)
				if i < len(v)-1 {
					sb.WriteString(`, `)
				}
			}

		case Escaped:
			sb.WriteString(dialects.EscapeCol(dialects.Dialect(dialect), string(v)))

		case []Escaped:
			for i, s := range v {
				sb.WriteString(dialects.EscapeCol(dialects.Dialect(dialect), string(s)))
				if i < len(v)-1 {
					sb.WriteString(`, `)
				}
			}

		case Repeatable:
			for i := 1; i <= int(v); i++ {
				sb.WriteString(dialects.Vars(dialects.Dialect(dialect), i))
				if i < int(v) {
					sb.WriteString(`, `)
				}
			}

		case int:
			sb.WriteString(dialects.Vars(dialects.Dialect(dialect), v))
		}
	}

	return sb.String()
}

func fields(prefix string, args []string) string {
	var sb strings.Builder

	for _, arg := range args {
		sb.WriteString(prefix)
		sb.WriteString(arg)
		sb.WriteString(`, `)
	}

	return sb.String()
}

func pls(args []Escaped, dialect string) string {
	var sb strings.Builder

	for i, arg := range args {
		sb.WriteString(dialects.EscapeCol(dialects.Dialect(dialect), string(arg)))
		sb.WriteString(" = ")
		sb.WriteString(dialects.Vars(dialects.Dialect(dialect), i+1))
		if i < len(args)-1 {
			sb.WriteString(`, `)
		}
	}

	return sb.String()
}
