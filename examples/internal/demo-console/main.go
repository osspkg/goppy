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
				Title: "Choose variant",
				Items: []string{
					"Kubernetes", "Docker", "Terraform", "Ansible",
					"Prometheus", "Grafana", "Vault", "Consul",
					"Nginx", "PostgreSQL", "Redis", "Kafka",
				},
				CallBack: func(args ...string) {
					fmt.Println("Selected:", args)
				},
				//MultiChoice: true,
				//MaxCols:     10,
			}

			for i := 0; i < 100; i++ {
				m.Items = append(m.Items, fmt.Sprintf("%d", i))
			}

			m.Run()

		})
	}))
	cmd.Exec()
}
