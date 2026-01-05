/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package algorithm_test

import (
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v3/auth/token/algorithm"
)

func TestUnit_Get(t *testing.T) {
	alg, err := algorithm.Get(algorithm.Name("a"))
	casecheck.Error(t, err)
	casecheck.Nil(t, alg)

	alg, err = algorithm.Get(algorithm.RS256)
	casecheck.NoError(t, err)
	casecheck.NotNil(t, alg)
}
