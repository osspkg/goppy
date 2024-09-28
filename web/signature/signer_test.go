/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package signature_test

import (
	"testing"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/goppy/v2/web/signature"
)

func TestUnit_Signature(t *testing.T) {
	sign := signature.NewSHA256("123", "456")

	body := []byte("hello")
	hash := "b7089b0463bf766946fc467102671dbe91659f17a7a19145cd68138c36b00555"

	casecheck.Equal(t, "123", sign.ID())
	casecheck.Equal(t, hash, sign.CreateString(body))
	casecheck.True(t, sign.Validate(body, hash))
}

func TestUnit_Storage(t *testing.T) {
	store := signature.NewStorage()

	store.Add(signature.NewSHA256("1", "0"))
	store.Add(signature.NewSHA256("2", "0"))
	store.Add(signature.NewSHA256("3", "0"))
	store.Add(signature.NewSHA256("5", "0"))
	casecheck.Equal(t, 4, store.Count())

	store.Add(signature.NewMD5("5", "0"))
	casecheck.Equal(t, 4, store.Count())

	casecheck.Nil(t, store.Get("4"))
	s := store.Get("5")
	casecheck.NotNil(t, s)
	casecheck.Equal(t, "5", s.ID())
	casecheck.Equal(t, "hmac-md5", s.Algorithm())
}
