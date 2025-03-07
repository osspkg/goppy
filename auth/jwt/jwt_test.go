/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jwt_test

import (
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v2/auth/jwt"
)

func TestUnit_ConfigJWT(t *testing.T) {
	conf := &jwt.ConfigGroup{}

	err := conf.Validate()
	casecheck.Error(t, err)

	conf.Default()

	err = conf.Validate()
	casecheck.NoError(t, err)
}
