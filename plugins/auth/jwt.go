package auth

import (
	"fmt"
	"time"

	"github.com/deweppro/go-sdk/auth/jwt"
	"github.com/deweppro/go-sdk/random"
	"github.com/deweppro/goppy/plugins"
)

type (
	// ConfigJWT jwt config model
	ConfigJWT struct {
		JWT []jwt.Config `yaml:"jwt"`
	}
)

func (v *ConfigJWT) Default() {
	if len(v.JWT) == 0 {
		for i := 0; i < 5; i++ {
			v.JWT = append(v.JWT, jwt.Config{
				ID:        random.String(6),
				Key:       random.String(32),
				Algorithm: jwt.AlgHS256,
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
		case jwt.AlgHS256, jwt.AlgHS384, jwt.AlgHS512:
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
			return jwt.New(conf.JWT)
		},
	}
}

type JWT interface {
	Sign(payload interface{}, ttl time.Duration) (string, error)
	Verify(token string, payload interface{}) (*jwt.Header, error)
}
