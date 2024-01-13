/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/syscall"
	"go.osspkg.com/goppy/ws/client"
	"go.osspkg.com/goppy/xlog"
)

func main() {
	group := iosync.NewGroup()
	ctx, cncl := context.WithCancel(context.TODO())
	cli := client.New(ctx, "ws://127.0.0.1:8088/ws", xlog.Default())
	defer cli.Close()
	go syscall.OnStop(func() {
		cli.Close()
	})
	group.Background(func() {
		err := cli.DialAndListen()
		if err != nil {
			xlog.WithError("err", err).Errorf("ws dial")
		}
	})
	<-time.After(100 * time.Millisecond)

	obs := client.NewObservable(cli)

	obs.Subscribe(1, []int{0}).
		Listen(func(arg client.ListenArg) {
			var vv json.RawMessage
			if err := arg.Decode(&vv); err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(vv))
		},
			client.PipeTake(1),
			client.PipeTimeout(1*time.Second),
		)

	obs.Subscribe(99, nil).
		Listen(func(arg client.ListenArg) {
			var vv json.RawMessage
			if err := arg.Decode(&vv); err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(vv))
		},
			client.PipeTake(3),
		)

	cncl()
	group.Wait()
}
