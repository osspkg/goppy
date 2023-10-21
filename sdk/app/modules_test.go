/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app_test

import (
	"testing"

	application "github.com/osspkg/goppy/sdk/app"
	"github.com/stretchr/testify/require"
)

func TestUnit_Modules(t *testing.T) {
	tmp1 := application.Modules{8, 9, "W"}
	tmp2 := application.Modules{18, 19, "aW", tmp1}
	main := application.Modules{1, 2, "qqq"}.Add(tmp2).Add(99)

	require.Equal(t, application.Modules{1, 2, "qqq", 18, 19, "aW", 8, 9, "W", 99}, main)
}
