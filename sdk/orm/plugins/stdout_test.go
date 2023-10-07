/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package plugins

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStdOut(t *testing.T) {
	w := &bytes.Buffer{}

	tl := &stdout{Writer: w}

	_, err := tl.Write([]byte("h4gbffke9"))
	require.NoError(t, err)
	tl.Metric("15gh7netd8", time.Minute)
	require.NoError(t, err)

	result := w.String()
	require.Contains(t, result, "h4gbffke9")
	require.Contains(t, result, "15gh7netd8: 1m0s")
}
