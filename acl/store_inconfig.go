/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl

import (
	"encoding/base64"
	"fmt"
	"sync"

	"go.osspkg.com/logx"
)

type (
	storeInConfig struct {
		data map[string][]byte
		mux  sync.RWMutex
	}

	ConfigInConfigStorage struct {
		ACL map[string]string `yaml:"acl_users"`
	}
)

func NewInConfigStorage(c *ConfigInConfigStorage) Storage {
	v := &storeInConfig{}

	v.data = make(map[string][]byte, len(c.ACL))
	for key, val := range c.ACL {
		b, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			logx.Error("base64 decoding failed", "err", err, "value", val)
			continue
		}
		v.data[key] = b
	}

	return v
}

func (v *storeInConfig) FindACL(uid string) ([]byte, error) {
	v.mux.RLock()
	defer v.mux.RUnlock()

	if acl, ok := v.data[uid]; ok {
		return acl, nil
	}

	return nil, fmt.Errorf("%s not exist", uid)
}

func (v *storeInConfig) ChangeACL(uid string, _ []byte) error {
	return fmt.Errorf("change %s not supported", uid)
}
