/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package algorithm

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

type RSA struct {
	Hash crypto.Hash
}

func (h *RSA) Sign(key *KeyAny, message []byte) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	privateKey, ok := key.Private.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not *rsa.PrivateKey")
	}

	if !h.Hash.Available() {
		return nil, fmt.Errorf("hash `%s` is unavailable", h.Hash.String())
	}

	w := h.Hash.New()
	if _, err := w.Write(message); err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}

	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, h.Hash, w.Sum(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	return sig, nil
}

func (h *RSA) Verify(key *KeyAny, message, sig []byte) error {
	if key == nil {
		return fmt.Errorf("key is nil")
	}

	publicKey, ok := key.Public.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("public key is not *rsa.PublicKey")
	}

	if !h.Hash.Available() {
		return fmt.Errorf("hash `%s` is unavailable", h.Hash.String())
	}

	w := h.Hash.New()
	if _, err := w.Write(message); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return rsa.VerifyPKCS1v15(publicKey, h.Hash, w.Sum(nil), sig)
}

func (h *RSA) Decode(key *KeyBytes) (*KeyAny, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	result := &KeyAny{}

	var err error
	if result.Private, err = x509.ParsePKCS1PrivateKey(key.Private); err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	if result.Public, err = x509.ParsePKCS1PublicKey(key.Public); err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return result, nil
}

func (h *RSA) DecodeBase64(key *KeyString) (*KeyAny, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	publicKey, err := base64.StdEncoding.DecodeString(key.Public)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	privateKey, err := base64.StdEncoding.DecodeString(key.Private)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	return h.Decode(&KeyBytes{privateKey, publicKey})
}

func (h *RSA) Encode(key *KeyAny) (*KeyBytes, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	privateKey, ok := key.Private.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not *rsa.PrivateKey")
	}

	publicKey, ok := key.Public.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not *rsa.PublicKey")
	}

	return &KeyBytes{x509.MarshalPKCS1PrivateKey(privateKey), x509.MarshalPKCS1PublicKey(publicKey)}, nil
}

func (h *RSA) EncodeBase64(key *KeyAny) (*KeyString, error) {
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

func (h *RSA) GenerateKeys() (*KeyAny, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	return &KeyAny{privateKey, &privateKey.PublicKey}, nil
}
