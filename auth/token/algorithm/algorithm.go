/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package algorithm

import (
	"crypto"
	_ "crypto/md5"    //nolint:gosec
	_ "crypto/sha1"   //nolint:gosec
	_ "crypto/sha256" //nolint:gosec
	_ "crypto/sha512" //nolint:gosec

	"go.osspkg.com/errors"
	"go.osspkg.com/syncing"
	_ "golang.org/x/crypto/blake2b"   //nolint:gosec
	_ "golang.org/x/crypto/blake2s"   //nolint:gosec
	_ "golang.org/x/crypto/md4"       //nolint:gosec,staticcheck
	_ "golang.org/x/crypto/ripemd160" //nolint:gosec,staticcheck
	_ "golang.org/x/crypto/sha3"      //nolint:gosec
)

var (
	ErrAlgoNotFound = errors.New("algorithm not found")
)

const (
	HS256 Name = "HS256"
	HS384 Name = "HS384"
	HS512 Name = "HS512"
	RS256 Name = "RS256"
	RS384 Name = "RS384"
	RS512 Name = "RS512"
	ES256 Name = "ES256"
	ES384 Name = "ES384"
	ES512 Name = "ES512"
	EdDSA Name = "EdDSA"
)

var list = syncing.NewMap[Name, Entity](20)

func init() {
	list.Set(HS256, &HMAC{crypto.SHA256})
	list.Set(HS384, &HMAC{crypto.SHA384})
	list.Set(HS512, &HMAC{crypto.SHA512})

	list.Set(RS256, &RSA{crypto.SHA256})
	list.Set(RS384, &RSA{crypto.SHA384})
	list.Set(RS512, &RSA{crypto.SHA512})

	list.Set(ES256, &ECDSA{crypto.SHA256, 32, 256})
	list.Set(ES384, &ECDSA{crypto.SHA384, 48, 384})
	list.Set(ES512, &ECDSA{crypto.SHA512, 66, 521})

	list.Set(EdDSA, &ED25519{})
}

func Register(name Name, s Entity) {
	list.Set(name, s)
}

func Get(name Name) (Entity, error) {
	alg, ok := list.Get(name)
	if !ok {
		return nil, errors.Wrapf(ErrAlgoNotFound, "name `%s`", name)
	}
	return alg, nil
}
