/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"

	"go.osspkg.com/goppy/v3/console"
)

func main() {
	cmd := console.New("", "")
	cmd.RootCommand(console.NewCommand(func(setter console.CommandSetter) {
		setter.ExecFunc(func() {

			m := console.InteractiveMenu{
				Title: "Выбирите вариант:",
				Items: []string{
					"MySQL", "PostgreSQL", "SQLite",
				},
				CallBack: func(args ...string) {
					fmt.Println("Выбран:", args)
				},
				MultiChoice: true,
			}

			//for i := 0; i < 100; i++ {
			//	m.Items = append(m.Items, fmt.Sprintf("Item%d", i))
			//}

			m.Run()

		})
	}))
	cmd.Exec()
}
