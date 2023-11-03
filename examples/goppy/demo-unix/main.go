/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"time"

	"go.osspkg.com/goppy"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/unixsocket"
)

func main() {
	app := goppy.New()
	app.Plugins(
		unixsocket.WithServer(),
		unixsocket.WithClient(),
	)
	app.Plugins(
		plugins.Plugin{
			Resolve: func(s unixsocket.Server, c unixsocket.Client, conf *unixsocket.Config) error {

				s.Command("demo", func(bytes []byte) ([]byte, error) {
					fmt.Println("<", string(bytes))
					return append(bytes, " world"...), nil
				})

				time.AfterFunc(time.Second*5, func() {
					cc, err := c.Create(conf.Path)
					if err != nil {
						panic(err)
					}

					b, err := cc.ExecString("demo", "hello")
					if err != nil {
						panic(err)
					}
					fmt.Println(">", string(b))

					b, err = cc.ExecString("demo", "hello")
					if err != nil {
						panic(err)
					}
					fmt.Println(">", string(b))
				})

				time.AfterFunc(time.Second*15, func() {
					cc, err := c.Create(conf.Path)
					if err != nil {
						panic(err)
					}

					b, err := cc.ExecString("demo", "hello")
					if err != nil {
						panic(err)
					}
					fmt.Println(">", string(b))
				})

				return nil
			},
		},
	)
	app.Run()
}
