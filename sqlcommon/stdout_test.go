/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package sqlcommon

import (
	"bytes"
	"testing"
	"time"

	"go.osspkg.com/goppy/xtest"
)

func TestStdOut(t *testing.T) {
	w := &bytes.Buffer{}

	tl := &stdout{Writer: w}

	_, err := tl.Write([]byte("h4gbffke9"))
	xtest.NoError(t, err)
	tl.Metric("15gh7netd8", time.Minute)
	xtest.NoError(t, err)

	result := w.String()
	xtest.Contains(t, result, "h4gbffke9")
	xtest.Contains(t, result, "15gh7netd8: 1m0s")
}
