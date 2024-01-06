/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package aesgcm_test

import (
	"testing"

	"go.osspkg.com/goppy/encryption/aesgcm"
	"go.osspkg.com/goppy/random"
	"go.osspkg.com/goppy/xtest"
)

func TestUnit_Codec(t *testing.T) {
	rndKey := random.String(32)
	message := []byte("Hello World!")

	c, err := aesgcm.New(rndKey)
	xtest.NoError(t, err)

	enc1, err := c.Encrypt(message)
	xtest.NoError(t, err)

	dec1, err := c.Decrypt(enc1)
	xtest.NoError(t, err)

	xtest.Equal(t, message, dec1)

	c, err = aesgcm.New(rndKey)
	xtest.NoError(t, err)

	enc2, err := c.Encrypt(message)
	xtest.NoError(t, err)

	xtest.NotEqual(t, enc1, enc2)

	dec2, err := c.Decrypt(enc1)
	xtest.NoError(t, err)

	xtest.Equal(t, message, dec2)

}
