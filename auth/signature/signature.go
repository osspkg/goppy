/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package signature

import (
	"crypto"
	"crypto/hmac"
	_ "crypto/md5"    //nolint:gosec
	_ "crypto/sha1"   //nolint:gosec
	_ "crypto/sha256" //nolint:gosec
	_ "crypto/sha512" //nolint:gosec
	"encoding/hex"
	"fmt"

	_ "golang.org/x/crypto/blake2b"   //nolint:gosec
	_ "golang.org/x/crypto/blake2s"   //nolint:gosec
	_ "golang.org/x/crypto/md4"       //nolint:gosec,staticcheck
	_ "golang.org/x/crypto/ripemd160" //nolint:gosec,staticcheck
	_ "golang.org/x/crypto/sha3"      //nolint:gosec
)

var _ Signature = (*sign)(nil)

type (
	sign struct {
		id     string
		alg    string
		secret []byte
		hash   crypto.Hash
	}

	// Signature interface
	Signature interface {
		ID() string
		Algorithm() string
		Create(message []byte) (string, error)
		Verify(message []byte, sig string) bool
	}
)

// ID signature
func (s *sign) ID() string {
	return s.id
}

// Algorithm getter
func (s *sign) Algorithm() string {
	return s.alg
}

func (s *sign) createSig(message []byte) ([]byte, error) {
	if !s.hash.Available() {
		return nil, fmt.Errorf("hash `%s` is unavailable", s.hash.String())
	}

	w := hmac.New(s.hash.New, s.secret)
	if _, err := w.Write(message); err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}
	return w.Sum(nil), nil
}

// Create getting hash as string
func (s *sign) Create(message []byte) (string, error) {
	h, err := s.createSig(message)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h), nil
}

// Verify signature
func (s *sign) Verify(message []byte, sig string) bool {
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}

	h, err := s.createSig(message)
	if err != nil {
		return false
	}

	return hmac.Equal(h, sigBytes)
}

// NewSHA1 create sign sha1
func NewSHA1(id, secret string) Signature {
	return NewCustomSignature(id, secret, "hmac-sha1", crypto.SHA1)
}

// NewSHA256 create sign sha256
func NewSHA256(id, secret string) Signature {
	return NewCustomSignature(id, secret, "hmac-sha256", crypto.SHA256)
}

// NewSHA512 create sign sha512
func NewSHA512(id, secret string) Signature {
	return NewCustomSignature(id, secret, "hmac-sha512", crypto.SHA512)
}

// NewCustomSignature create sign with custom hash function
func NewCustomSignature(id, secret, alg string, chash crypto.Hash) Signature {
	return &sign{
		id:     id,
		alg:    alg,
		secret: []byte(secret),
		hash:   chash,
	}
}
