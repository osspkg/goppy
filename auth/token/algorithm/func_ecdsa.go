/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package algorithm

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
)

type ECDSA struct {
	Hash      crypto.Hash
	KeySize   int
	CurveBits int
}

func (h *ECDSA) Sign(key *KeyAny, message []byte) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	privateKey, ok := key.Private.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not *ecdsa.PrivateKey")
	}

	if !h.Hash.Available() {
		return nil, fmt.Errorf("hash `%s` is unavailable", h.Hash.String())
	}

	w := h.Hash.New()
	if _, err := w.Write(message); err != nil {
		return nil, err
	}

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, w.Sum(nil))
	if err != nil {
		return nil, err
	}

	curveBits := privateKey.Curve.Params().BitSize
	if curveBits != h.CurveBits {
		return nil, fmt.Errorf("invalid curve bits")
	}

	keyBytes := curveBits / 8
	if curveBits%8 > 0 {
		keyBytes += 1
	}

	sig := make([]byte, 2*keyBytes)
	r.FillBytes(sig[0:keyBytes])
	s.FillBytes(sig[keyBytes:])

	return sig, nil
}

func (h *ECDSA) Verify(key *KeyAny, message, sig []byte) error {
	if key == nil {
		return fmt.Errorf("key is nil")
	}

	publicKey, ok := key.Public.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("public key is not *rsa.PublicKey")
	}

	if !h.Hash.Available() {
		return fmt.Errorf("hash `%s` is unavailable", h.Hash.String())
	}

	w := h.Hash.New()
	if _, err := w.Write(message); err != nil {
		return err
	}

	r := big.NewInt(0).SetBytes(sig[:h.KeySize])
	s := big.NewInt(0).SetBytes(sig[h.KeySize:])

	if !ecdsa.Verify(publicKey, w.Sum(nil), r, s) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (h *ECDSA) Decode(key *KeyBytes) (*KeyAny, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	curve, err := h.getCurve()
	if err != nil {
		return nil, fmt.Errorf("failed to get curve: %w", err)
	}

	privateKey, err := ecdsa.ParseRawPrivateKey(curve, key.Private)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey, err := ecdsa.ParseUncompressedPublicKey(curve, key.Public)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &KeyAny{privateKey, publicKey}, nil
}

func (h *ECDSA) DecodeBase64(key *KeyString) (*KeyAny, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	result := &KeyBytes{}

	var err error
	if result.Private, err = base64.StdEncoding.DecodeString(key.Private); err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	if result.Public, err = base64.StdEncoding.DecodeString(key.Public); err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	return h.Decode(result)
}

func (h *ECDSA) Encode(key *KeyAny) (*KeyBytes, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	privateKey, ok := key.Private.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not *ecdsa.PrivateKey")
	}

	publicKey, ok := key.Public.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not *ecdsa.PublicKey")
	}

	result := &KeyBytes{}

	var err error
	if result.Private, err = privateKey.Bytes(); err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	if result.Public, err = publicKey.Bytes(); err != nil {
		return nil, fmt.Errorf("failed to encode public key: %w", err)
	}

	return result, nil

}

func (h *ECDSA) EncodeBase64(key *KeyAny) (*KeyString, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	keyBytes, err := h.Encode(key)
	if err != nil {
		return nil, err
	}

	return &KeyString{
		Private: base64.StdEncoding.EncodeToString(keyBytes.Private),
		Public:  base64.StdEncoding.EncodeToString(keyBytes.Public),
	}, nil
}

func (h *ECDSA) getCurve() (elliptic.Curve, error) {
	switch h.CurveBits {
	case 256:
		return elliptic.P256(), nil
	case 384:
		return elliptic.P384(), nil
	case 521:
		return elliptic.P521(), nil
	default:
		return nil, fmt.Errorf("invalid curve bits")
	}
}

func (h *ECDSA) GenerateKeys() (*KeyAny, error) {
	curve, err := h.getCurve()
	if err != nil {
		return nil, fmt.Errorf("failed to get curve: %w", err)
	}

	var privateKey *ecdsa.PrivateKey
	privateKey, err = ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	return &KeyAny{privateKey, &privateKey.PublicKey}, nil
}
