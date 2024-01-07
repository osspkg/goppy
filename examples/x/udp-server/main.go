/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"net"

	"go.osspkg.com/goppy/syscall"
	"go.osspkg.com/goppy/udp/server"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

func main() {
	serv := server.New(xlog.Default(), ":11111")
	serv.HandleFunc(&Echo{})
	fmt.Println(serv.Up(xc.New()))
	syscall.OnStop(func() {
		fmt.Println(serv.Down())
	})
}

type Echo struct{}

func (v *Echo) HandlerUDP(w server.Writer, addr net.Addr, b []byte) {
	fmt.Println(addr.String(), ">  ", string(b))
	w.WriteTo(b, addr) //nolint: errcheck
}
