/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package routine_test

import (
	"fmt"
	"testing"

	"go.osspkg.com/goppy/routine"
)

func TestUnit_Parallel(t *testing.T) {
	routine.Parallel(
		func() {
			fmt.Println("a")
		}, func() {
			fmt.Println("b")
		}, func() {
			fmt.Println("c")
		},
	)
}
