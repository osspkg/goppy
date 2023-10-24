/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.osspkg.com/goppy/sdk/syscall"

	"go.osspkg.com/goppy/sdk/iosync"
	"go.osspkg.com/goppy/sdk/log"

	"go.osspkg.com/goppy/sdk/netutil/websocket"
)

func main() {
	group := iosync.NewGroup()
	ctx, cncl := context.WithCancel(context.TODO())
	cli := websocket.NewClient(ctx, "ws://127.0.0.1:8088/ws", log.Default())
	defer cli.Close()
	go syscall.OnStop(func() {
		cli.Close()
	})
	group.Background(func() {
		err := cli.DialAndListen()
		if err != nil {
			log.WithError("err", err).Errorf("ws dial")
		}
	})
	<-time.After(100 * time.Millisecond)

	obs := websocket.NewObservable(cli)

	obs.Subscribe(1, []int{0}).
		Listen(func(arg websocket.ListenArg) {
			var vv json.RawMessage
			if err := arg.Decode(&vv); err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(vv))
		},
			websocket.PipeTake(1),
			websocket.PipeTimeout(1*time.Second),
		)

	obs.Subscribe(99, nil).
		Listen(func(arg websocket.ListenArg) {
			var vv json.RawMessage
			if err := arg.Decode(&vv); err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(vv))
		},
			websocket.PipeTake(3),
		)

	cncl()
	group.Wait()
}
