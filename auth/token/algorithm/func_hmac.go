/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package algorithm

import (
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"fmt"

	"go.osspkg.com/errors"
	"go.osspkg.com/random"
)

const hmacKeyMinLen = 32

type HMAC struct {
	Hash crypto.Hash
}

func (h *HMAC) Sign(key *KeyAny, message []byte) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	privateKey, ok := key.Private.([]byte)
	if !ok {
		return nil, errors.New("private key is not []byte")
	}

	if !h.Hash.Available() {
		return nil, fmt.Errorf("hash `%s` is unavailable", h.Hash.String())
	}

	w := hmac.New(h.Hash.New, privateKey)
	if _, err := w.Write(message); err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}
	return w.Sum(nil), nil
}

func (h *HMAC) Verify(key *KeyAny, message, sig []byte) error {
	if key == nil {
		return fmt.Errorf("key is nil")
	}

	sign, err := h.Sign(key, message)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	if !hmac.Equal(sign, sig) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (h *HMAC) Decode(key *KeyBytes) (*KeyAny, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	if len(key.Private) < hmacKeyMinLen {
		return nil, errors.New("private key is less 32 bytes")
	}

	return &KeyAny{key.Private, key.Private}, nil
}

func (h *HMAC) DecodeBase64(key *KeyString) (*KeyAny, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	privateKey, err := base64.StdEncoding.DecodeString(key.Private)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	return h.Decode(&KeyBytes{privateKey, privateKey})
}

func (h *HMAC) Encode(key *KeyAny) (*KeyBytes, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	privateKey, ok := key.Private.([]byte)
	if !ok {
		return nil, errors.New("private key is not []byte")
	}

	if len(privateKey) < hmacKeyMinLen {
		return nil, errors.New("private key is less 32 bytes")
	}

	return &KeyBytes{privateKey, privateKey}, nil
}

func (h *HMAC) EncodeBase64(key *KeyAny) (*KeyString, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	keyBytes, err := h.Encode(key)
	if err != nil {
		return nil, err
	}

	s := base64.StdEncoding.EncodeToString(keyBytes.Private)

	return &KeyString{s, s}, nil
}

func (h *HMAC) GenerateKeys() (*KeyAny, error) {
	privateKey := random.CryptoBytes(hmacKeyMinLen)

	return &KeyAny{privateKey, privateKey}, nil
}
