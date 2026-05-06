/*
 *  Copyright (c) 2021-2023 Mikhail Knyazhev <markus621@gmail.com>. All rights reserved.
 *  Use of this source code is governed by a BSD-3-Clause license that can be found in the LICENSE file.
 */

package assets_test

import (
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v3/pkg/assets"
)

func TestUnit_ReadDir(t *testing.T) {
	c := assets.New()

	casecheck.NoError(t, c.FromDir("."))
	casecheck.Equal(t, c.List(), []string{
		"/cache.go",
		"/files.go",
		"/files_test.go",
		"/targz.go",
	})
}
