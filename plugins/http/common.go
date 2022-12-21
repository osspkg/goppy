package http

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"

	"github.com/deweppro/go-http/pkg/signature"
)

type Signature interface {
	ID() string
	Algorithm() string
	Create(b []byte) []byte
	CreateString(b []byte) string
	Validate(b []byte, ex string) bool
}

// NewSHA1 create sign sha1
func NewSHA1(id, secret string) Signature {
	return signature.NewCustomSignature(id, secret, "hmac-sha1", sha1.New)
}

// NewSHA256 create sign sha256
func NewSHA256(id, secret string) Signature {
	return signature.NewCustomSignature(id, secret, "hmac-sha256", sha256.New)
}

// NewSHA512 create sign sha512
func NewSHA512(id, secret string) Signature {
	return signature.NewCustomSignature(id, secret, "hmac-sha512", sha512.New)
}
