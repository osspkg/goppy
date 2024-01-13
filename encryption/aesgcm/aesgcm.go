/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package aesgcm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

const keySize = 32

type Codec struct {
	key   []byte
	block cipher.Block
}

func New(key string) (*Codec, error) {
	kb := []byte(key)
	if len(kb) < keySize {
		return nil, fmt.Errorf("invalid key len")
	}
	obj := &Codec{
		key: kb[:keySize],
	}
	block, err := aes.NewCipher(obj.key)
	if err != nil {
		return nil, err
	}
	obj.block = block
	return obj, nil
}

func (v *Codec) Encrypt(plaintext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(v.block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func (v *Codec) Decrypt(ciphertext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(v.block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("invalid message len")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
