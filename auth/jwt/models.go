/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jwt

//go:generate easyjson

//easyjson:json
type Header struct {
	Kid       string `json:"kid"`
	Alg       string `json:"alg"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"eat"`
}
