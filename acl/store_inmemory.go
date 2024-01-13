/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl

import (
	"sync"
)

type OptionInMemoryStorage func(v *storeInMemory)

func OptionInMemoryStorageSetupData(data map[string]string) OptionInMemoryStorage {
	return func(v *storeInMemory) {
		v.data = make(map[string]string, len(data))
		for key, val := range data {
			v.data[key] = val
		}
	}
}

type storeInMemory struct {
	data map[string]string
	mux  sync.Mutex
}

func NewInMemoryStorage(opts ...OptionInMemoryStorage) Storage {
	v := &storeInMemory{
		data: make(map[string]string),
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

func (v *storeInMemory) FindACL(email string) (string, error) {
	v.mux.Lock()
	defer v.mux.Unlock()

	if acl, ok := v.data[email]; ok {
		return acl, nil
	}
	return "", errUserNotFound
}

func (v *storeInMemory) ChangeACL(email, data string) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.data[email] = data
	return nil
}
