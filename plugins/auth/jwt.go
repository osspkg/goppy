package auth

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

	"github.com/deweppro/go-sdk/random"
	"github.com/deweppro/goppy/plugins"
)

const (
	JWTAlgHS256 = "HS256"
	JWTAlgHS384 = "HS384"
	JWTAlgHS512 = "HS512"
)

type (
	// ConfigJWT jwt config model
	ConfigJWT struct {
		JWT ConfigJWTItem `yaml:"jwt"`
	}

	ConfigJWTItem struct {
		Key       string        `yaml:"key"`
		TTL       time.Duration `yaml:"ttl"`
		Algorithm string        `yaml:"alg"`
		Cookie    string        `yaml:"cookie"`
	}
)

func (v *ConfigJWT) Default() {
	if len(v.JWT.Key) == 0 {
		v.JWT = ConfigJWTItem{
			Key:       random.String(32),
			TTL:       time.Hour * 24,
			Algorithm: JWTAlgHS256,
			Cookie:    "jwt",
		}
	}
}

func (v *ConfigJWT) Validate() error {
	if len(v.JWT.Key) < 32 {
		return fmt.Errorf("jwt key less than 32 characters")
	}
	if len(v.JWT.Cookie) == 0 {
		return fmt.Errorf("jwt cookie name is empty")
	}
	switch v.JWT.Algorithm {
	case JWTAlgHS256, JWTAlgHS384, JWTAlgHS512:
	default:
		return fmt.Errorf("jwt algorithm not supported")
	}
	return nil
}

// WithJWT init jwt provider
func WithJWT() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigJWT{},
		Inject: func(conf *ConfigJWT) (JWT, error) {
			return newJWT(conf.JWT)
		},
	}
}

//easyjson:json
type JWTHeader struct {
	Alg       string `json:"alg"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"eat"`
}

type (
	JWT interface {
		Sign(payload interface{}) (string, error)
		Extend(token string) (string, error)
		Verify(token string, payload interface{}) (*JWTHeader, error)
		CookieName() string
	}

	_jwt struct {
		conf ConfigJWTItem
		hash func() hash.Hash
		key  []byte
	}
)

func newJWT(c ConfigJWTItem) (JWT, error) {
	var h func() hash.Hash

	switch c.Algorithm {
	case JWTAlgHS256:
		h = sha256.New
	case JWTAlgHS384:
		h = sha512.New384
	case JWTAlgHS512:
		h = sha512.New
	default:
		return nil, fmt.Errorf("jwt algorithm not supported")
	}

	return &_jwt{conf: c, hash: h, key: []byte(c.Key)}, nil
}

func (v *_jwt) calcHash(data []byte) ([]byte, error) {
	mac := hmac.New(v.hash, v.key)
	if _, err := mac.Write(data); err != nil {
		return nil, err
	}
	result := mac.Sum(nil)
	return result, nil
}

func (v *_jwt) Sign(payload interface{}) (string, error) {
	h, err := (&JWTHeader{
		Alg:       v.conf.Algorithm,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(v.conf.TTL).Unix(),
	}).MarshalJSON()
	if err != nil {
		return "", err
	}
	result := base64.StdEncoding.EncodeToString(h)

	p, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	result += "." + base64.StdEncoding.EncodeToString(p)

	s, err := v.calcHash([]byte(result))
	if err != nil {
		return "", err
	}
	result += "." + base64.StdEncoding.EncodeToString(s)

	return result, nil
}

func (v *_jwt) Extend(token string) (string, error) {
	data := strings.Split(token, ".")
	if len(data) != 3 {
		return "", fmt.Errorf("invalid jwt format")
	}

	h, err := base64.StdEncoding.DecodeString(data[0])
	if err != nil {
		return "", err
	}
	header := &JWTHeader{}
	if err = header.UnmarshalJSON(h); err != nil {
		return "", err
	}

	header.ExpiresAt = time.Now().Add(v.conf.TTL).Unix()

	h, err = header.MarshalJSON()
	if err != nil {
		return "", err
	}
	data[0] = base64.StdEncoding.EncodeToString(h)

	sig, err := v.calcHash([]byte(data[0] + "." + data[1]))
	if err != nil {
		return "", err
	}

	data[2] = base64.StdEncoding.EncodeToString(sig)

	return strings.Join(data, "."), nil
}

func (v *_jwt) Verify(token string, payload interface{}) (*JWTHeader, error) {
	data := strings.Split(token, ".")
	if len(data) != 3 {
		return nil, fmt.Errorf("invalid jwt format")
	}

	h, err := base64.StdEncoding.DecodeString(data[0])
	if err != nil {
		return nil, err
	}
	header := &JWTHeader{}
	if err = header.UnmarshalJSON(h); err != nil {
		return nil, err
	}

	if header.Alg != v.conf.Algorithm {
		return nil, fmt.Errorf("invalid jwt algorithm")
	}
	if header.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("jwt expired")
	}

	expected, err := base64.StdEncoding.DecodeString(data[2])
	if err != nil {
		return nil, err
	}
	actual, err := v.calcHash([]byte(data[0] + "." + data[1]))
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

	if err = json.Unmarshal(p, payload); err != nil {
		return nil, err
	}

	return header, nil
}

func (v *_jwt) CookieName() string {
	return v.conf.Cookie
}
