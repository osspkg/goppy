/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app_test

import (
	"testing"

	"go.osspkg.com/goppy/app"
	"go.osspkg.com/goppy/xtest"
)

func TestUnit_Modules(t *testing.T) {
	tmp1 := app.Modules{8, 9, "W"}
	tmp2 := app.Modules{18, 19, "aW", tmp1}
	main := app.Modules{1, 2, "qqq"}.Add(tmp2).Add(99)

	xtest.Equal(t, app.Modules{1, 2, "qqq", 18, 19, "aW", 8, 9, "W", 99}, main)
}
