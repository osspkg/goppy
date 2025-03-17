/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jwt

import (
	"fmt"

	"github.com/google/uuid"
	"go.osspkg.com/random"
)

type (
	ConfigGroup struct {
		JWT Config `yaml:"jwt"`
	}

	Config struct {
		Option Option `yaml:"option"`
		Keys   []Key  `yaml:"keys"`
	}

	Key struct {
		ID        string `yaml:"id"`
		Key       string `yaml:"key"`
		Algorithm string `yaml:"alg"`
	}

	Option struct {
		HeaderName bool   `yaml:"header_name"`
		CookieName string `yaml:"cookie_name"`
	}
)

func (v *ConfigGroup) Default() {
	if len(v.JWT.Keys) == 0 {
		for i := 0; i < 10; i++ {
			v.JWT.Keys = append(v.JWT.Keys, Key{
				ID:        uuid.NewString(),
				Key:       random.String(32),
				Algorithm: AlgHS256,
			})
		}
	}
}

func (v *ConfigGroup) Validate() error {
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
