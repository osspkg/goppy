/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jwt

//go:generate easyjson

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"strings"
	"time"

	"go.osspkg.com/goppy/sdk/encryption/aesgcm"
)

const (
	AlgHS256 = "HS256"
	AlgHS384 = "HS384"
	AlgHS512 = "HS512"
)

type Config struct {
	ID        string `yaml:"id"`
	Key       string `yaml:"key"`
	Algorithm string `yaml:"alg"`
}

//easyjson:json
type Header struct {
	Kid       string `json:"kid"`
	Alg       string `json:"alg"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"eat"`
}

type (
	JWT struct {
		pool map[string]*Pool
	}

	Pool struct {
		conf  Config
		hash  func() hash.Hash
		key   []byte
		codec *aesgcm.Codec
	}
)

func New(conf []Config) (*JWT, error) {
	obj := &JWT{pool: make(map[string]*Pool)}

	for _, c := range conf {
		var h func() hash.Hash
		switch c.Algorithm {
		case AlgHS256:
			h = sha256.New
		case AlgHS384:
			h = sha512.New384
		case AlgHS512:
			h = sha512.New
		default:
			return nil, fmt.Errorf("jwt algorithm not supported")
		}
		codec, err := aesgcm.New(c.Key)
		if err != nil {
			return nil, fmt.Errorf("jwt init codec: %w", err)
		}
		obj.pool[c.ID] = &Pool{conf: c, hash: h, key: []byte(c.Key), codec: codec}
	}

	return obj, nil
}

func (v *JWT) rndPool() (*Pool, error) {
	for _, p := range v.pool {
		return p, nil
	}
	return nil, fmt.Errorf("jwt pool is empty")
}

func (v *JWT) getPool(id string) (*Pool, error) {
	p, ok := v.pool[id]
	if ok {
		return p, nil
	}
	return nil, fmt.Errorf("jwt pool not found")
}

func (v *JWT) calcHash(hash func() hash.Hash, key []byte, data []byte) ([]byte, error) {
	mac := hmac.New(hash, key)
	if _, err := mac.Write(data); err != nil {
		return nil, err
	}
	result := mac.Sum(nil)
	return result, nil
}

func (v *JWT) Sign(payload interface{}, ttl time.Duration) (string, error) {
	pool, err := v.rndPool()
	if err != nil {
		return "", err
	}

	rh := &Header{
		Kid:       pool.conf.ID,
		Alg:       pool.conf.Algorithm,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}
	h, err := json.Marshal(rh)
	if err != nil {
		return "", err
	}
	result := base64.StdEncoding.EncodeToString(h)

	p, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	p, err = pool.codec.Encrypt(p)
	if err != nil {
		return "", err
	}
	result += "." + base64.StdEncoding.EncodeToString(p)

	s, err := v.calcHash(pool.hash, pool.key, []byte(result))
	if err != nil {
		return "", err
	}
	result += "." + base64.StdEncoding.EncodeToString(s)

	return result, nil
}

func (v *JWT) Verify(token string, payload interface{}) (*Header, error) {
	data := strings.Split(token, ".")
	if len(data) != 3 {
		return nil, fmt.Errorf("invalid jwt format")
	}

	h, err := base64.StdEncoding.DecodeString(data[0])
	if err != nil {
		return nil, err
	}
	header := &Header{}
	if err = json.Unmarshal(h, header); err != nil {
		return nil, err
	}

	pool, err := v.getPool(header.Kid)
	if err != nil {
		return nil, err
	}

	if header.Alg != pool.conf.Algorithm {
		return nil, fmt.Errorf("invalid jwt algorithm")
	}
	if header.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("jwt expired")
	}

	expected, err := base64.StdEncoding.DecodeString(data[2])
	if err != nil {
		return nil, err
	}
	actual, err := v.calcHash(pool.hash, pool.key, []byte(data[0]+"."+data[1]))
	if err != nil {
		return nil, err
	}
	if !hmac.Equal(expected, actual) {
		return nil, fmt.Errorf("invalid jwt signature")
	}

	p, err := base64.StdEncoding.DecodeString(data[1])
	if err != nil {
		return nil, err
	}
	p, err = pool.codec.Decrypt(p)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(p, payload); err != nil {
		return nil, err
	}

	return header, nil
}
