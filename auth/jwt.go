/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package auth

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
	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/web"
	"go.osspkg.com/random"
)

//go:generate easyjson

const (
	AlgHS256 = "HS256"
	AlgHS384 = "HS384"
	AlgHS512 = "HS512"
)

type (
	// ConfigJWT jwt config model
	ConfigJWT struct {
		JWT JWTConfig `yaml:"jwt"`
	}

	JWTConfig struct {
		Option JWTOption `yaml:"option"`
		Keys   []JWTKey  `yaml:"keys"`
	}

	JWTKey struct {
		ID        string `yaml:"id"`
		Key       string `yaml:"key"`
		Algorithm string `yaml:"alg"`
	}

	//easyjson:json
	JWTHeader struct {
		Kid       string `json:"kid"`
		Alg       string `json:"alg"`
		IssuedAt  int64  `json:"iat"`
		ExpiresAt int64  `json:"eat"`
	}

	JWTOption struct {
		HeaderName bool   `yaml:"header_name"`
		CookieName string `yaml:"cookie_name"`
	}
)

func (v *ConfigJWT) Default() {
	if len(v.JWT.Keys) == 0 {
		for i := 0; i < 10; i++ {
			v.JWT.Keys = append(v.JWT.Keys, JWTKey{
				ID:        random.String(8),
				Key:       random.String(32),
				Algorithm: AlgHS256,
			})
		}
	}
}

func (v *ConfigJWT) Validate() error {
	if len(v.JWT.Keys) == 0 {
		return fmt.Errorf("jwt keys config is empty")
	}
	for _, vv := range v.JWT.Keys {
		if len(vv.ID) == 0 {
			return fmt.Errorf("jwt key id is empty")
		}
		if len(vv.Key) != 32 {
			return fmt.Errorf("jwt key less than 32 characters")
		}
		switch vv.Algorithm {
		case AlgHS256, AlgHS384, AlgHS512:
		default:
			return fmt.Errorf("jwt algorithm not supported")
		}
	}

	return nil
}

type jwtPool struct {
	conf  JWTKey
	hash  func() hash.Hash
	key   []byte
	codec *aesgcm.Codec
}

func (v *jwtPool) hashing(data []byte) ([]byte, error) {
	mac := hmac.New(v.hash, v.key)
	if _, err := mac.Write(data); err != nil {
		return nil, err
	}
	result := mac.Sum(nil)
	return result, nil
}

func (v *jwtPool) encrypt(payload interface{}) ([]byte, error) {
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

func (v *jwtPool) decrypt(data []byte, payload interface{}) error {
	b, err := v.codec.Decrypt(data)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, payload); err != nil {
		return err
	}
	return nil
}

// WithJWT init jwt provider
func WithJWT() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigJWT{},
		Inject: func(conf *ConfigJWT) JWT {
			return &jwtService{
				conf: conf.JWT,
				pool: make(map[string]*jwtPool),
			}
		},
	}
}

type JWT interface {
	GuardMiddleware() web.Middleware
	Sign(payload interface{}, ttl time.Duration) (string, error)
	SignCookie(ctx web.Context, payload interface{}, ttl time.Duration) error
}

type jwtService struct {
	conf JWTConfig
	pool map[string]*jwtPool
}

func (v *jwtService) Up() error {
	for _, c := range v.conf.Keys {
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
		v.pool[c.ID] = &jwtPool{conf: c, hash: h, key: []byte(c.Key), codec: codec}
	}
	return nil
}

func (v *jwtService) Down() error {
	return nil
}

func (v *jwtService) randomPool() (*jwtPool, error) {
	for _, p := range v.pool {
		return p, nil
	}
	return nil, fmt.Errorf("jwt pool is empty")
}

func (v *jwtService) getPoolById(id string) (*jwtPool, error) {
	p, ok := v.pool[id]
	if ok {
		return p, nil
	}
	return nil, fmt.Errorf("jwt pool not found")
}

func (v *jwtService) signPayload(payload interface{}, ttl time.Duration) (string, error) {
	pool, err := v.randomPool()
	if err != nil {
		return "", err
	}

	rh := &JWTHeader{
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

	p, err := pool.encrypt(payload)
	if err != nil {
		return "", err
	}
	result += "." + base64.StdEncoding.EncodeToString(p)

	s, err := pool.hashing([]byte(result))
	if err != nil {
		return "", err
	}
	result += "." + base64.StdEncoding.EncodeToString(s)

	return result, nil
}

func (v *jwtService) verifyPayload(token string, payload interface{}) (*JWTHeader, error) {
	data := strings.Split(token, ".")
	if len(data) != 3 {
		return nil, fmt.Errorf("invalid jwt format")
	}

	h, err := base64.StdEncoding.DecodeString(data[0])
	if err != nil {
		return nil, err
	}

	header := &JWTHeader{}
	if err = json.Unmarshal(h, header); err != nil {
		return nil, err
	}

	pool, err := v.getPoolById(header.Kid)
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

	actual, err := pool.hashing([]byte(data[0] + "." + data[1]))
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

	if err = pool.decrypt(p, payload); err != nil {
		return nil, err
	}

	return header, nil
}

func (v *jwtService) Sign(payload interface{}, ttl time.Duration) (string, error) {
	return v.signPayload(payload, ttl)
}

func (v *jwtService) SignCookie(ctx web.Context, payload interface{}, ttl time.Duration) error {
	tok, err := v.signPayload(payload, ttl)
	if err != nil {
		return err
	}
	ctx.Cookie().Set(&http.Cookie{
		Name:     v.conf.Option.CookieName,
		Value:    tok,
		Path:     "/",
		Domain:   ctx.URL().Host,
		Expires:  time.Now().Add(ttl),
		Secure:   true,
		HttpOnly: true,
	})
	return nil
}

func (v *jwtService) GuardMiddleware() web.Middleware {
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			val := ""

			if v.conf.Option.HeaderName {
				val = r.Header.Get("Authorization")
				if len(val) > 7 && strings.HasPrefix(val, "Bearer ") {
					val = val[6:]
				}
			}

			if len(val) == 0 && len(v.conf.Option.CookieName) > 0 {
				cv, err := r.Cookie(v.conf.Option.CookieName)
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
