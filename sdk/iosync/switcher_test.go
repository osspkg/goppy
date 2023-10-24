/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package iosync_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.osspkg.com/goppy/sdk/iosync"
)

func TestNewSwitch(t *testing.T) {
	sync := iosync.NewSwitch()

	require.False(t, sync.IsOn())
	require.True(t, sync.IsOff())

	require.True(t, sync.On())
	require.False(t, sync.On())

	require.False(t, sync.IsOff())
	require.True(t, sync.IsOn())

}
