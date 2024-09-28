/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"bytes"
	"testing"

	"go.osspkg.com/casecheck"
)

func TestNewMetric(t *testing.T) {
	w := &bytes.Buffer{}

	demo1 := NewMetric(nil)
	demo1.ExecutionTime("hello1", func() {})

	demo2 := NewMetric(StdOutWriter)
	demo2.ExecutionTime("hello2", func() {})

	result := w.String()
	casecheck.NotContains(t, result, "hello1")
	casecheck.Contains(t, result, "hello2")
}
