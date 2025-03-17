/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jwt

import (
	"crypto/hmac"
	"encoding/json"
	"hash"

	"go.osspkg.com/encrypt/aesgcm"
)

type keyPool struct {
	conf  Key
	hash  func() hash.Hash
	key   []byte
	codec *aesgcm.Codec
}

func (v *keyPool) hashing(data []byte) ([]byte, error) {
	mac := hmac.New(v.hash, v.key)
	if _, err := mac.Write(data); err != nil {
		return nil, err
	}
	result := mac.Sum(nil)
	return result, nil
}

func (v *keyPool) encrypt(payload interface{}) ([]byte, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	p, err = v.codec.Encrypt(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (v *keyPool) decrypt(data []byte, payload interface{}) error {
	b, err := v.codec.Decrypt(data)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, payload); err != nil {
		return err
	}
	return nil
}
