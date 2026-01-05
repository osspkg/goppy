/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package console

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"
)

var helpTemplate = `{{if len .Description | ne 0}}NAME
	{{.Name}} - {{.Description}}
{{end}}SYNOPSIS
	{{.Name}} {{.Curr}} {{.Args}}
{{if len .CurrDesc | ne 0}}DESCRIPTION
	{{.CurrDesc}}
{{end}}{{if len .Flags | ne 0}}ARGUMENTS
{{range $ex := .Flags}}	{{$ex}}
{{end}}{{end}}{{if len .Next | ne 0}}COMMANDS
{{range $ex := .Next}}	{{$ex}}
{{end}}{{end}}
`

type helpModel struct {
	Name        string
	Description string
	ShowCommand bool

	Args  string
	Flags []string

	Curr     string
	CurrDesc string
	Next     []string
}

func helpView(tool string, desc string, c CommandGetter, global []CommandGetter, args []string) {
	model := &helpModel{
		ShowCommand: c != nil,
		Name:        tool,
		Description: desc,

		Curr: strings.Join(args, " ") + " " + c.Name(),
		CurrDesc: func() string {
			if c == nil {
				return ""
			}
			return c.Description()
		}(),
		Next: func() (out []string) {
			if c == nil {
				return
			}
			var chars int
			next := c.List()
			for _, v := range next {
				if chars < len(v.Name()) {
					chars = len(v.Name())
				}
			}
			sort.Slice(next, func(i, j int) bool {
				return next[i].Name() < next[j].Name()
			})
			chars += 3
			for _, v := range next {
				out = append(out,
					v.Name()+
						strings.Repeat(" ", chars-len(v.Name()))+
						v.Description())
			}

			return
		}(),
	}

	if c != nil {
		model.Args = "[arg]"
		model.Flags = func() (out []string) {
			chars := 0
			for _, all := range append([]CommandGetter{c}, global...) {
				all.Flags().Info(func(_ bool, name string, _ interface{}, _ string) {
					length := len(name)
					if length > 2 {
						length += 2
					} else {
						length++
					}
					if length > chars {
						chars = length
					}
				})
			}
			chars += 2
			for _, gc := range global {
				gc.Flags().Info(func(req bool, name string, value interface{}, usage string) {
					defaultValue, i := "", 1
					if !req {
						defaultValue = fmt.Sprintf("(default: %+v)", value)
					}
					if len(name) > 1 {
						i = 2
					}
					out = append(out, fmt.Sprintf(
						"%s%s%s    %s %s [GLOBAL]",
						strings.Repeat("-", i),
						name,
						strings.Repeat(" ", chars-len(name)-i),
						usage,
						defaultValue,
					))
				})
			}
			c.Flags().Info(func(req bool, name string, value interface{}, usage string) {
				defaultValue, i := "", 1
				if !req {
					defaultValue = fmt.Sprintf("(default: %+v)", value)
				}
				if len(name) > 1 {
					i = 2
				}
				out = append(out, fmt.Sprintf(
					"%s%s%s    %s %s",
					strings.Repeat("-", i),
					name,
					strings.Repeat(" ", chars-len(name)-i),
					usage,
					defaultValue,
				))
			})
			return out
		}()
	}

	if err := template.Must(template.New("").Parse(helpTemplate)).Execute(os.Stdout, model); err != nil {
		Fatalf(err.Error())
	}
	os.Exit(0)
}
