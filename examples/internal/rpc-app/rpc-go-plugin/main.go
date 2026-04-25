/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"context"
	"fmt"
)

func main() {

}

func Start(ctx context.Context, opts map[string]string) error {
	return nil
}

func Stop() error {
	return nil
}

func Call(ctx context.Context, method string, params, result any) (err error) {
	return fmt.Errorf("call ok")
}
