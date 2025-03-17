/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"net/http"
	"strings"
	"time"

	"go.osspkg.com/encrypt/aesgcm"

	"go.osspkg.com/goppy/v2/web"
)

type JWT interface {
	GuardMiddleware() web.Middleware
	Sign(payload interface{}, ttl time.Duration) (string, error)
	SignCookie(ctx web.Context, payload interface{}, ttl time.Duration) error
}

func New(c Config) JWT {
	return &service{
		config:  c,
		keyPool: make(map[string]*keyPool),
	}
}

type service struct {
	config  Config
	keyPool map[string]*keyPool
}

func (v *service) Up() error {
	for _, c := range v.config.Keys {
		var h func() hash.Hash
		switch c.Algorithm {
		case AlgHS256:
			h = sha256.New
		case AlgHS384:
			h = sha512.New384
		case AlgHS512:
			h = sha512.New
		default:
			return fmt.Errorf("jwt algorithm not supported in `%s`", c.ID)
		}

		codec, err := aesgcm.New([]byte(c.Key))
		if err != nil {
			return fmt.Errorf("jwt init codec: %w", err)
		}

		v.keyPool[c.ID] = &keyPool{conf: c, hash: h, key: []byte(c.Key), codec: codec}
	}

	return nil
}

func (v *service) Down() error {
	return nil
}

func (v *service) randomKP() (*keyPool, error) {
	for _, p := range v.keyPool {
		return p, nil
	}

	return nil, fmt.Errorf("jwt keyPool is empty")
}

func (v *service) getKPById(id string) (*keyPool, error) {
	if p, ok := v.keyPool[id]; ok {
		return p, nil
	}

	return nil, fmt.Errorf("jwt key not found")
}

func (v *service) signPayload(payload interface{}, ttl time.Duration) (string, error) {
	kp, err := v.randomKP()
	if err != nil {
		return "", err
	}

	header := &Header{
		Kid:       kp.conf.ID,
		Alg:       kp.conf.Algorithm,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}

	h, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	result := dataPool.Get()
	defer func() { dataPool.Put(result) }()

	result.WriteString(base64.StdEncoding.EncodeToString(h)) //nolint:errcheck

	p, err := kp.encrypt(payload)
	if err != nil {
		return "", err
	}

	result.WriteString(".")                                  //nolint:errcheck
	result.WriteString(base64.StdEncoding.EncodeToString(p)) //nolint:errcheck

	s, err := kp.hashing(result.Bytes())
	if err != nil {
		return "", err
	}

	result.WriteString(".")                                  //nolint:errcheck
	result.WriteString(base64.StdEncoding.EncodeToString(s)) //nolint:errcheck

	return result.String(), nil
}

func (v *service) verifyPayload(token string, payload interface{}) (*Header, error) {
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

	kp, err := v.getKPById(header.Kid)
	if err != nil {
		return nil, err
	}

	if header.Alg != kp.conf.Algorithm {
		return nil, fmt.Errorf("invalid jwt algorithm")
	}

	if header.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("jwt expired")
	}

	expected, err := base64.StdEncoding.DecodeString(data[2])
	if err != nil {
		return nil, err
	}

	actual, err := kp.hashing([]byte(data[0] + "." + data[1]))
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

	if err = kp.decrypt(p, payload); err != nil {
		return nil, err
	}

	return header, nil
}

func (v *service) Sign(payload interface{}, ttl time.Duration) (string, error) {
	return v.signPayload(payload, ttl)
}

func (v *service) SignCookie(ctx web.Context, payload interface{}, ttl time.Duration) error {
	tok, err := v.signPayload(payload, ttl)
	if err != nil {
		return err
	}
	ctx.Cookie().Set(&http.Cookie{
		Name:     v.config.Option.CookieName,
		Value:    tok,
		Path:     "/",
		Domain:   ctx.URL().Host,
		Expires:  time.Now().Add(ttl),
		Secure:   true,
		HttpOnly: true,
	})
	return nil
}

func (v *service) GuardMiddleware() web.Middleware {
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			val := ""

			if v.config.Option.HeaderName {
				val = r.Header.Get("Authorization")
				if len(val) > 7 && strings.HasPrefix(val, "Bearer ") {
					val = val[6:]
				}
			}

			if len(val) == 0 && len(v.config.Option.CookieName) > 0 {
				cv, err := r.Cookie(v.config.Option.CookieName)
				if err == nil && len(cv.Value) > 0 {
					val = cv.Value
				}
			}

			if len(val) == 0 {
				http.Error(w, "authorization required", http.StatusUnauthorized)
				return
			}

			var raw json.RawMessage

			if _, err := v.verifyPayload(val, &raw); err != nil {
				http.Error(w, "authorization required", http.StatusUnauthorized)
				return
			}

			ctx = setJWTPayloadContext(ctx, raw)

			call(w, r.WithContext(ctx))
		}
	}
}
