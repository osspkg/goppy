/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package aesgcm_test

import (
	"testing"

	"github.com/osspkg/goppy/sdk/encryption/aesgcm"
	"github.com/osspkg/goppy/sdk/random"
	"github.com/stretchr/testify/require"
)

func TestUnit_Codec(t *testing.T) {
	rndKey := random.String(32)
	message := []byte("Hello World!")

	c, err := aesgcm.New(rndKey)
	require.NoError(t, err)

	enc1, err := c.Encrypt(message)
	require.NoError(t, err)

	dec1, err := c.Decrypt(enc1)
	require.NoError(t, err)

	require.Equal(t, message, dec1)

	c, err = aesgcm.New(rndKey)
	require.NoError(t, err)

	enc2, err := c.Encrypt(message)
	require.NoError(t, err)

	require.NotEqual(t, enc1, enc2)

	dec2, err := c.Decrypt(enc1)
	require.NoError(t, err)

	require.Equal(t, message, dec2)

}
