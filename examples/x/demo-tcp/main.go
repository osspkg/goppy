/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"os"
	"time"

	"go.osspkg.com/goppy/syscall"
	"go.osspkg.com/goppy/tcp/server"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

func main() {
	conf := server.ConfigItem{
		Pools:   []server.Pool{{Port: 11111, Certs: nil}},
		Timeout: 5 * time.Second,
	}
	l := xlog.New()
	l.SetLevel(xlog.LevelDebug)
	l.SetFormatter(xlog.NewFormatString())
	l.SetOutput(os.Stdout)
	defer l.Close()

	s := server.New(conf, l)
	c := xc.New()

	fmt.Println(s.Up(c))
	syscall.OnStop(func() {
		c.Close()
	})
	fmt.Println(s.Down())
}
