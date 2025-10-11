/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package algorithm_test

import (
	"fmt"
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v2/auth/token/algorithm"
)

func TestUnit_HS256(t *testing.T) {
	alg, err := algorithm.Get(algorithm.HS256)
	casecheck.NoError(t, err)
	casecheck.NotNil(t, alg)

	key, err := alg.GenerateKeys()
	casecheck.NoError(t, err)
	casecheck.NotNil(t, key)

	strKey, err := alg.EncodeBase64(key)
	casecheck.NoError(t, err)
	casecheck.NotNil(t, strKey)

	fmt.Println(strKey.Private)
	fmt.Println(strKey.Public)

	strKey = &algorithm.KeyString{
		Private: "xvUpRhgYyRhi2Wx0A593AzmkuQ9bEzooXGX+gjJHKdk=",
		Public:  "xvUpRhgYyRhi2Wx0A593AzmkuQ9bEzooXGX+gjJHKdk=",
	}

	key, err = alg.DecodeBase64(strKey)
	casecheck.NoError(t, err)
	casecheck.NotNil(t, key)

	msg := []byte("hello world")

	sign, err := alg.Sign(key, msg)
	casecheck.NoError(t, err)

	casecheck.NoError(t, alg.Verify(key, msg, sign))
}
