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
		JWT []ConfigJWTItem `yaml:"jwt"`
	}

	ConfigJWTItem struct {
		ID        string `yaml:"id"`
		Key       string `yaml:"key"`
		Algorithm string `yaml:"alg"`
	}
)

func (v *ConfigJWT) Default() {
	if len(v.JWT) == 0 {
		for i := 0; i < 5; i++ {
			v.JWT = append(v.JWT, ConfigJWTItem{
				ID:        random.String(6),
				Key:       random.String(32),
				Algorithm: JWTAlgHS256,
			})
		}
	}
}

func (v *ConfigJWT) Validate() error {
	if len(v.JWT) == 0 {
		return fmt.Errorf("jwt config is empty")
	}
	for _, vv := range v.JWT {
		if len(vv.ID) == 0 {
			return fmt.Errorf("jwt key id is empty")
		}
		if len(vv.Key) < 32 {
			return fmt.Errorf("jwt key less than 32 characters")
		}
		switch vv.Algorithm {
		case JWTAlgHS256, JWTAlgHS384, JWTAlgHS512:
		default:
			return fmt.Errorf("jwt algorithm not supported")
		}
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
	Kid       string `json:"kid"`
	Alg       string `json:"alg"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"eat"`
}

type (
	JWT interface {
		Sign(payload interface{}, ttl time.Duration) (string, error)
		Verify(token string, payload interface{}) (*JWTHeader, error)
	}

	_jwt struct {
		pool map[string]*_jwtPoolItem
	}

	_jwtPoolItem struct {
		conf ConfigJWTItem
		hash func() hash.Hash
		key  []byte
	}
)

func newJWT(conf []ConfigJWTItem) (JWT, error) {
	obj := &_jwt{pool: make(map[string]*_jwtPoolItem)}

	for _, c := range conf {
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
		obj.pool[c.ID] = &_jwtPoolItem{conf: c, hash: h, key: []byte(c.Key)}
	}

	return obj, nil
}

func (v *_jwt) randPool() (*_jwtPoolItem, error) {
	for _, p := range v.pool {
		return p, nil
	}
	return nil, fmt.Errorf("jwt pool is empty")
}

func (v *_jwt) getPool(id string) (*_jwtPoolItem, error) {
	p, ok := v.pool[id]
	if ok {
		return p, nil
	}
	return nil, fmt.Errorf("jwt pool not found")
}

func (v *_jwt) calcHash(hash func() hash.Hash, key []byte, data []byte) ([]byte, error) {
	mac := hmac.New(hash, key)
	if _, err := mac.Write(data); err != nil {
		return nil, err
	}
	result := mac.Sum(nil)
	return result, nil
}

func (v *_jwt) Sign(payload interface{}, ttl time.Duration) (string, error) {
	pool, err := v.randPool()
	if err != nil {
		return "", err
	}

	h, err := (&JWTHeader{
		Kid:       pool.conf.ID,
		Alg:       pool.conf.Algorithm,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(ttl).Unix(),
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

	s, err := v.calcHash(pool.hash, pool.key, []byte(result))
	if err != nil {
		return "", err
	}
	result += "." + base64.StdEncoding.EncodeToString(s)

	return result, nil
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

	if err = json.Unmarshal(p, payload); err != nil {
		return nil, err
	}

	return header, nil
}
