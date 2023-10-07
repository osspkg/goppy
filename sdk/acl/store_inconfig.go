/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl

type (
	storeInConfig struct {
		data map[string]string
	}

	ConfigInConfigStorage struct {
		ACL map[string]string `yaml:"acl_users"`
	}
)

func NewInConfigStorage(c *ConfigInConfigStorage) Storage {
	v := &storeInConfig{}

	v.data = make(map[string]string, len(c.ACL))
	for key, val := range c.ACL {
		v.data[key] = val
	}

	return v
}

func (v *storeInConfig) FindACL(email string) (string, error) {
	if acl, ok := v.data[email]; ok {
		return acl, nil
	}
	return "", errUserNotFound
}

func (v *storeInConfig) ChangeACL(email, data string) error {
	return errChangeNotSupported
}
