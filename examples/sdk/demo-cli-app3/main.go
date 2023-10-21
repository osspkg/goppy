/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"strings"

	"github.com/osspkg/goppy/sdk/console"
)

func main() {
	console.ShowDebug(true)

	app := console.New("tool", "help tool")

	cmd := console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("a", "command a")
		setter.ExecFunc(func(args []string) {
			fmt.Println("a", args)
		})

		setter.AddCommand(console.NewCommand(func(setter console.CommandSetter) {
			setter.Setup("b", "command b")
			setter.ExecFunc(func(args []string) {
				fmt.Println("b", args)
			})
		}))

	})

	root := console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("root", "command root")
		setter.Flag(func(setter console.FlagsSetter) {
			setter.Bool("aaa", "bool a")
		})
		setter.ArgumentFunc(func(s []string) ([]string, error) {
			return []string{strings.Join(s, "-")}, nil
		})
		setter.ExecFunc(func(args []string, a bool) {
			fmt.Println("root", args, a)
		})
	})

	app.RootCommand(root)
	app.AddCommand(cmd)
	app.Exec()
}
