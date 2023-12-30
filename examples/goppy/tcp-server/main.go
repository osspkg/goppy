/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"

	"go.osspkg.com/goppy"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/tcp"
)

func main() {
	app := goppy.New()
	app.Plugins(
		tcp.WithServer(),
		plugins.Plugin{
			Inject: func(server tcp.Server) error {
				h := &Echo{}
				server.HandleFunc(h)
				return nil
			},
		},
	)
	app.Run()
}

type Echo struct{}

func (e *Echo) HandlerTCP(w tcp.Response, r tcp.Request) {
	for {
		b, err := r.ReadLine()
		if err != nil {
			fmt.Println("ERR:", r.Addr().String(), err)
			return
		}

		fmt.Println("GET:", r.Addr().String())
		fmt.Println(string(b))

		fmt.Fprintf(w, "ECHO\n")
	}
}
