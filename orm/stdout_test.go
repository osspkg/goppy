/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"bytes"
	"testing"
	"time"

	"go.osspkg.com/casecheck"
)

func TestStdOut(t *testing.T) {
	w := &bytes.Buffer{}

	tl := &stdout{Writer: w}

	_, err := tl.Write([]byte("h4gbffke9"))
	casecheck.NoError(t, err)
	tl.Metric("15gh7netd8", time.Minute)
	casecheck.NoError(t, err)

	result := w.String()
	casecheck.Contains(t, result, "h4gbffke9")
	casecheck.Contains(t, result, "15gh7netd8: 1m0s")
}
