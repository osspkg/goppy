/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package sqlcommon

import (
	"bytes"
	"testing"

	"go.osspkg.com/goppy/xtest"
)

func TestNewMetric(t *testing.T) {
	w := &bytes.Buffer{}
	tl := &stdout{Writer: w}

	demo1 := NewMetric(nil)
	demo1.ExecutionTime("hello1", func() {})

	demo2 := NewMetric(tl)
	demo2.ExecutionTime("hello2", func() {})

	result := w.String()
	xtest.NotContains(t, result, "hello1")
	xtest.Contains(t, result, "hello2")
}
