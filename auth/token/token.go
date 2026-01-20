/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package token

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.osspkg.com/errors"
	"go.osspkg.com/ioutils/cache"

	"go.osspkg.com/goppy/v3/auth/token/algorithm"
	"go.osspkg.com/goppy/v3/auth/token/internal/b64"
	"go.osspkg.com/goppy/v3/auth/token/internal/byteops"
	"go.osspkg.com/goppy/v3/web"
)

var (
	separator = []byte(".")
)

type Token interface {
	Audience() string
	CookieName() string
	HeaderName() string
	SecureOnly() bool

	CreateJWT(payload json.Marshaler, audience string, lifetime time.Duration) (uuid.UUID, []byte, error)
	VerifyJWT(token []byte) (head *Header, payload []byte, err error)

	CreateCookie(ctx web.Ctx, payload json.Marshaler, lifetime time.Duration) (uuid.UUID, error)
	FlushCookie(ctx web.Ctx)
}

type (
	_service struct {
		opt ConfigOption
		jwt cache.Cache[string, *signerEntity]
	}

	signerEntity struct {
		ID        string
		Type      Type
		Issuer    string
		Algorithm algorithm.Name
		KeyEntity *algorithm.KeyAny
	}
)

func New(c Config) (Token, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	v := &_service{
		opt: c.Option,
		jwt: cache.New[string, *signerEntity](),
	}

	err := errors.Wrap(
		v.setupSigners(c.Sign),
	)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (v *_service) Audience() string {
	return v.opt.Audience
}

func (v *_service) CookieName() string {
	return v.opt.CookieName
}

func (v *_service) HeaderName() string {
	return v.opt.HeaderName
}

func (v *_service) SecureOnly() bool {
	return v.opt.SecureOnly
}

func (v *_service) setupSigners(c ConfigSign) error {
	for _, k := range c.Keys {
		alg, err := algorithm.Get(k.Algo)
		if err != nil {
			return fmt.Errorf("jwt: get algorithm %s: %w", k.Algo, err)
		}

		keyBytes := &algorithm.KeyBytes{}
		if keyBytes.Private, err = k.Key.getBytes(); err != nil {
			return fmt.Errorf("jwt: get key bytes: %w", err)
		}
		if keyBytes.Public, err = k.Cert.getBytes(); err != nil {
			return fmt.Errorf("jwt: get cert bytes: %w", err)
		}

		key, err := alg.Decode(keyBytes)
		if err != nil {
			return fmt.Errorf("jwt: decode keys: %w", err)
		}

		v.jwt.Set(k.ID, &signerEntity{
			ID:        k.ID,
			Type:      c.Type,
			Issuer:    c.Issuer,
			Algorithm: k.Algo,
			KeyEntity: key,
		})
	}

	if v.jwt.Size() <= 0 {
		return fmt.Errorf("jwt: no signers configured")
	}

	return nil
}

func (v *_service) CreateJWT(payload json.Marshaler, audience string, lifetime time.Duration) (uuid.UUID, []byte, error) {
	if payload == nil {
		return uuid.Nil, nil, fmt.Errorf("jwt: empty payload")
	}

	_, sig, ok := v.jwt.One()
	if !ok {
		return uuid.Nil, nil, fmt.Errorf("jwt: failed getting signer")
	}

	alg, err := algorithm.Get(sig.Algorithm)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("jwt: get algorithm %s: %w", sig.Algorithm, err)
	}

	tokId, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("jwt: failed creating token id: %w", err)
	}

	w := dataPool.Get()
	defer func() { dataPool.Put(w) }()

	header := &Header{
		Algorithm: sig.Algorithm,
		TokenType: sig.Type,
		TokenID:   tokId.String(),
		Issuer:    sig.Issuer,
		Audience:  audience,
		KeyID:     sig.ID,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(lifetime).Unix(),
	}

	bh, err := json.Marshal(header)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("jwt: failed to marshal header: %w", err)
	}
	if _, err = w.Write(b64.UrlEncode(bh)); err != nil {
		return uuid.Nil, nil, fmt.Errorf("jwt: failed to write header: %w", err)
	}

	if _, err = w.Write(separator); err != nil {
		return uuid.Nil, nil, fmt.Errorf("jwt: failed to write separator: %w", err)
	}

	bp, err := payload.MarshalJSON()
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("jwt: failed to marshal payload: %w", err)
	}
	if _, err = w.Write(b64.UrlEncode(bp)); err != nil {
		return uuid.Nil, nil, fmt.Errorf("jwt: failed to write payload: %w", err)
	}

	bs, err := alg.Sign(sig.KeyEntity, w.Bytes())
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("jwt: failed to sign payload: %w", err)
	}

	if _, err = w.Write(separator); err != nil {
		return uuid.Nil, nil, fmt.Errorf("jwt: failed to write separator: %w", err)
	}

	if _, err = w.Write(b64.UrlEncode(bs)); err != nil {
		return uuid.Nil, nil, fmt.Errorf("jwt: failed to write payload: %w", err)
	}

	return tokId, w.Bytes(), nil
}

func (v *_service) VerifyJWT(token []byte) (*Header, []byte, error) {
	if len(token) == 0 {
		return nil, nil, ErrEmptyToken
	}

	index := byteops.Indexes(token, separator[0])
	if len(index) != 2 || index[1]+1 > len(token)-1 {
		return nil, nil, fmt.Errorf("jwt: invalid token format")
	}

	header := &Header{}
	if err := json.Unmarshal(b64.UrlDecode(token[0:index[0]]), header); err != nil {
		return nil, nil, fmt.Errorf("jwt: failed to unmarshal header: %w", err)
	}
	if header.TokenType != TypeJWT {
		return nil, nil, fmt.Errorf("jwt: invalid token type")
	}
	currTime := time.Now().Unix()
	if header.IssuedAt > currTime {
		return nil, nil, fmt.Errorf("jwt: invalid token issued")
	}
	if header.ExpiresAt < currTime {
		return nil, nil, fmt.Errorf("jwt: invalid token expired")
	}
	alg, err := algorithm.Get(header.Algorithm)
	if err != nil {
		return nil, nil, fmt.Errorf("jwt: get algorithm %s: %w", header.Algorithm, err)
	}
	sig, ok := v.jwt.Get(header.KeyID)
	if !ok {
		return nil, nil, fmt.Errorf("jwt: invalid token key id")
	}

	if err = alg.Verify(
		sig.KeyEntity,
		token[0:index[1]],
		b64.UrlDecode(token[index[1]+1:]),
	); err != nil {
		return nil, nil, fmt.Errorf("jwt: failed to verify signature: %w", err)
	}

	payload := b64.UrlDecode(token[index[0]+1 : index[1]])
	if len(payload) == 0 {
		return nil, nil, fmt.Errorf("jwt: invalid token payload")
	}

	return header, payload, nil
}

func (v *_service) CreateCookie(ctx web.Ctx, payload json.Marshaler, lifetime time.Duration) (uuid.UUID, error) {
	if payload == nil {
		return uuid.Nil, fmt.Errorf("jwt: empty payload")
	}

	tokId, tok, err := v.CreateJWT(payload, v.Audience(), lifetime)
	if err != nil {
		return uuid.Nil, fmt.Errorf("jwt: failed to create token: %w", err)
	}

	ctx.Cookie().Set(&http.Cookie{
		Name:     v.opt.CookieName,
		Value:    string(tok),
		Path:     "/",
		Domain:   ctx.URL().Host,
		Expires:  time.Now().Add(lifetime),
		Secure:   v.opt.SecureOnly,
		HttpOnly: true,
	})

	return tokId, nil
}

func (v *_service) FlushCookie(ctx web.Ctx) {
	ctx.Cookie().Set(&http.Cookie{
		Name:     v.opt.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   ctx.URL().Host,
		Expires:  time.Now().Add(-3600 * time.Hour),
		Secure:   v.opt.SecureOnly,
		HttpOnly: true,
	})
}
