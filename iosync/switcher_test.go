/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package iosync_test

import (
	"testing"

	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/xtest"
)

func TestNewSwitch(t *testing.T) {
	sync := iosync.NewSwitch()

	xtest.False(t, sync.IsOn())
	xtest.True(t, sync.IsOff())

	xtest.True(t, sync.On())
	xtest.False(t, sync.On())

	xtest.False(t, sync.IsOff())
	xtest.True(t, sync.IsOn())

}
