/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package signature_test

import (
	"crypto"
	"testing"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/random"

	"go.osspkg.com/goppy/v3/plugins/auth/signature"
)

func TestUnit_Signature(t *testing.T) {
	sign := signature.NewSHA256("123", "456")

	body := []byte("hello")
	sigExpected := "DHpeeYNf0LqPGa+AQXoFphlxVk7CA2xQZP3oAuJTjvA="
	nonce := "gjsjggjyyjetyjznvsjkuy"

	casecheck.Equal(t, "123", sign.ID())
	sigActual, err := sign.Create(body, nonce)
	t.Log(sigActual)
	casecheck.NoError(t, err)
	casecheck.Equal(t, sigExpected, sigActual)
	casecheck.True(t, sign.Verify(body, nonce, sigActual))
}

func Benchmark_Signature(b *testing.B) {
	sign := signature.NewSHA256("123", random.String(crypto.SHA256.Size()))
	body := []byte("hello")
	nonce := "qazxswedcvfrtgb"

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sign.Create(body, nonce)
		}
	})
}
