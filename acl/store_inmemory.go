/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl

import (
	"fmt"

	"go.osspkg.com/ioutils/cache"
)

type storeInMemory struct {
	data cache.Cache[string, []byte]
}

func NewInMemoryStorage(data map[string][]byte) Storage {
	v := &storeInMemory{
		data: cache.New[string, []byte](),
	}

	v.data.Replace(data)

	return v
}

func (v *storeInMemory) FindACL(uid string) ([]byte, error) {
	b, ok := v.data.Get(uid)
	if !ok {
		return nil, fmt.Errorf("%s not exist", uid)
	}

	return b, nil
}

func (v *storeInMemory) ChangeACL(uid string, access []byte) error {
	v.data.Set(uid, access)

	return nil
}
