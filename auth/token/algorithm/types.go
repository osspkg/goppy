/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package algorithm

type Name string

type Entity interface {
	GenerateKeys() (*KeyAny, error)

	Sign(key *KeyAny, message []byte) ([]byte, error)
	Verify(key *KeyAny, message, sig []byte) error

	Decode(*KeyBytes) (*KeyAny, error)
	Encode(*KeyAny) (*KeyBytes, error)

	DecodeBase64(*KeyString) (*KeyAny, error)
	EncodeBase64(*KeyAny) (*KeyString, error)
}

type KeyAny struct {
	Private any
	Public  any
}

type KeyBytes struct {
	Private []byte
	Public  []byte
}

type KeyString struct {
	Private string
	Public  string
}
