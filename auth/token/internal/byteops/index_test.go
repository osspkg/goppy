/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package byteops_test

import (
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v3/auth/token/internal/byteops"
)

func TestUnit_Indexes(t *testing.T) {
	data := []byte("1.2.3.4")
	inx := byteops.Indexes(data, '.')
	casecheck.Equal(t, []int{1, 3, 5}, inx)
}
