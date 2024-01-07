/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"context"
	"fmt"
	"time"

	"go.osspkg.com/goppy/routine"
	"go.osspkg.com/goppy/syscall"
	"go.osspkg.com/goppy/udp/client"
)

func main() {
	cli, err := client.New("127.0.0.1:11111")
	if err != nil {
		panic(err)
	}
	cli.HandleFunc(&Printer{})
	routine.Interval(context.TODO(), time.Second, func(ctx context.Context) {
		if _, err = cli.Write([]byte("123")); err != nil {
			fmt.Println(err)
		}
	})
	syscall.OnStop(func() {
		fmt.Println(cli.Close())
	})
}

type Printer struct{}

func (v *Printer) HandlerUDP(err error, b []byte) {
	fmt.Println(err, ">", len(b), "...", string(b))
}
