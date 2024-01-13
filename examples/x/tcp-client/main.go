/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"go.osspkg.com/goppy/tcp"
)

func main() {
	cli, err := tcp.NewClient(tcp.ClientConfig{
		Address:           "0.0.0.0:8080",
		Timeout:           1 * time.Second,
		ServerMaxBodySize: 5e+10,
	})
	if err != nil {
		panic(err)
	}
	for i := 0; i < 3; i++ {
		b, err := cli.Do(context.TODO(), bytes.NewBuffer([]byte(fmt.Sprintf("Hello %d", i))))
		fmt.Println(err)
		fmt.Println(string(b))
		time.Sleep(6 * time.Second)
	}
}
