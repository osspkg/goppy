/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package random_test

import (
	"bytes"
	"fmt"
	"testing"

	"go.osspkg.com/goppy/random"
)

func TestUnit_Bytes(t *testing.T) {
	max := 10
	r1 := random.Bytes(max)
	r2 := random.Bytes(max)

	fmt.Println(string(r1), string(r2))

	if len(r1) != max || len(r2) != max {
		t.Errorf("invalid len, is not %d", max)
	}
	if bytes.Equal(r1, r2) {
		t.Errorf("result is not random")
	}
}

func TestUnit_BytesOf(t *testing.T) {
	max := 10
	src := []byte("1234567890")
	r1 := random.BytesOf(max, src)
	r2 := random.BytesOf(max, src)

	fmt.Println(string(r1), string(r2))

	if len(r1) != max || len(r2) != max {
		t.Errorf("invalid len, is not %d", max)
	}
	if bytes.Equal(r1, r2) {
		t.Errorf("result is not random")
	}
}

func Benchmark_Bytes64(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		random.Bytes(64)
	}
}

func Benchmark_Bytes256(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		random.Bytes(256)
	}
}
