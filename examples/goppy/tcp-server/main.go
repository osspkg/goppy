/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"io"

	"go.osspkg.com/goppy"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/tcp"
)

func main() {
	app := goppy.New()
	app.Plugins(
		tcp.WithServer(),
		plugins.Plugin{
			Inject: func(server tcp.ServerTCP) error {
				h := &Echo{}
				server.HandleFunc(h)
				server.ErrHandleFunc(h)
				return nil
			},
		},
	)
	app.Run()
}

type Echo struct{}

func (e *Echo) HandlerTCP(c tcp.Connect) {
	fmt.Println("IN", c.Addr(), c.ID())
	b, err := io.ReadAll(c)
	if err != nil {
		fmt.Println("ERR:", err)
		return
	}
	fmt.Println(string(b))
	fmt.Fprintf(c, "HTTP/2.0 200 OK\n\r\n")
	//fmt.Fprintf(c, "HTTP/2.0 200 OK\nConnection: close\n\r\n")
	//c.Close()
}

func (e *Echo) ErrHandlerTCP(c tcp.ErrConnect) {
	fmt.Fprintf(c, "HTTP/1.1 500 OK\n\n%s\r\n", c.Err().Error())
}
