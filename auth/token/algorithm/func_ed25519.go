/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package algorithm

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"go.osspkg.com/errors"
)

type ED25519 struct{}

func (h *ED25519) Sign(key *KeyAny, message []byte) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	privateKey, ok := key.Private.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not ed25519.PrivateKey")
	}

	return ed25519.Sign(privateKey, message), nil
}

func (h *ED25519) Verify(key *KeyAny, message, sig []byte) error {
	if key == nil {
		return fmt.Errorf("key is nil")
	}

	publicKey, ok := key.Public.(ed25519.PublicKey)
	if !ok {
		return errors.New("public key is not ed25519.PublicKey")
	}

	if len(publicKey) != ed25519.PublicKeySize {
		return errors.New("invalid public key size")
	}

	if !ed25519.Verify(publicKey, message, sig) {
		return errors.New("invalid signature")
	}

	return nil
}

func (h *ED25519) Decode(key *KeyBytes) (*KeyAny, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	return &KeyAny{
		Private: ed25519.PrivateKey(key.Private),
		Public:  ed25519.PublicKey(key.Public),
	}, nil
}

func (h *ED25519) DecodeBase64(key *KeyString) (*KeyAny, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	result := &KeyBytes{}

	var err error
	if result.Public, err = base64.StdEncoding.DecodeString(key.Public); err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	if result.Private, err = base64.StdEncoding.DecodeString(key.Private); err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	return h.Decode(result)
}

func (h *ED25519) Encode(key *KeyAny) (*KeyBytes, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	result := &KeyBytes{}

	var ok bool
	if result.Private, ok = key.Private.(ed25519.PrivateKey); !ok {
		return nil, errors.New("private key is not ed25519.PrivateKey")
	}
	if result.Public, ok = key.Public.(ed25519.PublicKey); !ok {
		return nil, errors.New("public key is not ed25519.PublicKey")
	}

	return result, nil
}

func (h *ED25519) EncodeBase64(key *KeyAny) (*KeyString, error) {
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

func (h *ED25519) GenerateKeys() (*KeyAny, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	return &KeyAny{privateKey, publicKey}, nil
}
