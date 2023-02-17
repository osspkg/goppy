package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/deweppro/goppy/plugins/web"
)

const (
	jwtPayload = "jwtp"
	jwtHeader  = "jwth"
)

type jwtContext string

type ctx interface {
	Context() context.Context
}

func setJWTPayload(ctx context.Context, value []byte) context.Context {
	return context.WithValue(ctx, jwtContext(jwtPayload), value)
}

func GetJWTPayload(c ctx, payload interface{}) error {
	value, ok := c.Context().Value(jwtContext(jwtPayload)).([]byte)
	if !ok {
		return fmt.Errorf("jwt payload not found")
	}
	return json.Unmarshal(value, payload)
}

func setJWTHeader(ctx context.Context, value *JWTHeader) context.Context {
	return context.WithValue(ctx, jwtContext(jwtHeader), *value)
}

func GetJWTHeader(c ctx, payload interface{}) *JWTHeader {
	value, ok := c.Context().Value(jwtContext(jwtPayload)).(JWTHeader)
	if !ok {
		return nil
	}
	return &value
}

func JWTGuardMiddleware(j JWT) web.Middleware {
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			c, err := r.Cookie(j.CookieName())
			if err != nil {
				http.Error(w, "authorization required", http.StatusUnauthorized)
				return
			}

			var raw json.RawMessage
			h, err := j.Verify(c.Value, &raw)
			if err != nil {
				http.Error(w, "authorization required", http.StatusUnauthorized)
				return
			}

			ctx = setJWTHeader(ctx, h)
			ctx = setJWTPayload(ctx, raw)

			call(w, r.WithContext(ctx))
		}
	}
}
