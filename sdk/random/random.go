/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package random

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	digest = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-+=~*@#$%&?!<>")
)

func BytesOf(n int, src []byte) []byte {
	tmp := make([]byte, len(src))
	copy(tmp, src)
	rand.Shuffle(len(tmp), func(i, j int) {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	})
	b := make([]byte, n)
	for i := range b {
		b[i] = tmp[rand.Intn(len(tmp))]
	}
	return b
}

func StringOf(n int, src string) string {
	return string(BytesOf(n, []byte(src)))
}

func Bytes(n int) []byte {
	return BytesOf(n, digest)
}

func String(n int) string {
	return string(Bytes(n))
}
