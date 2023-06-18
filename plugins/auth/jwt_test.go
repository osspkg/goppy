/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package auth_test

import (
	"testing"

	"github.com/osspkg/goppy/plugins/auth"
	"github.com/stretchr/testify/require"
)

func TestUnit_ConfigJWT(t *testing.T) {
	conf := &auth.ConfigJWT{}

	err := conf.Validate()
	require.Error(t, err)

	conf.Default()

	err = conf.Validate()
	require.NoError(t, err)
}
