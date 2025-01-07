/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl

import (
	"sync"
	"time"
)

type (
	cache struct {
		size uint
		data map[string]*item
		mux  sync.Mutex
	}
	item struct {
		Val []uint8
		Ts  int64
	}
)

func newCache(size uint) *cache {
	return &cache{
		size: size,
		data: make(map[string]*item),
	}
}

func (v *cache) Has(email string) bool {
	v.mux.Lock()
	defer v.mux.Unlock()

	_, ok := v.data[email]

	return ok
}

func (v *cache) Get(email string, feature uint16) (uint8, error) {
	v.mux.Lock()
	defer v.mux.Unlock()

	access, ok := v.data[email]
	if !ok {
		return 0, errUserNotFound
	}

	if feature > uint16(v.size-1) {
		return 0, errFeatureGreaterMax
	}

	access.Ts = time.Now().Unix()
	return access.Val[feature], nil
}

func (v *cache) GetAll(email string) ([]uint8, error) {
	v.mux.Lock()
	defer v.mux.Unlock()

	access, ok := v.data[email]
	if !ok {
		return nil, errUserNotFound
	}

	access.Ts = ttl()

	tmp := make([]uint8, v.size)
	for i, level := range access.Val {
		if uint(i) >= v.size {
			break
		}
		tmp[i] = validateLevel(level)
	}

	return tmp, nil
}

func (v *cache) Set(email string, feature uint16, level uint8) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	if feature > uint16(v.size-1) {
		return errFeatureGreaterMax
	}

	access, ok := v.data[email]
	if !ok {
		access = &item{Val: make([]uint8, v.size)}
		v.data[email] = access
	}

	access.Ts = ttl()
	access.Val[feature] = validateLevel(level)
	return nil
}

func (v *cache) SetAll(email string, levels ...uint8) {
	v.mux.Lock()
	defer v.mux.Unlock()

	access, ok := v.data[email]
	if !ok {
		access = &item{Val: make([]uint8, v.size)}
		v.data[email] = access
	}

	access.Ts = ttl()
	for i, level := range levels {
		if uint(i) >= v.size {
			break
		}
		access.Val[i] = validateLevel(level)
	}
}

func (v *cache) Flush(email string) {
	v.mux.Lock()
	defer v.mux.Unlock()

	delete(v.data, email)
}

func (v *cache) FlushByTime(ts int64) {
	v.mux.Lock()
	defer v.mux.Unlock()

	for email, access := range v.data {
		if access.Ts < ts {
			delete(v.data, email)
		}
	}
}

func (v *cache) List() []string {
	v.mux.Lock()
	defer v.mux.Unlock()

	tmp := make([]string, 0, len(v.data))
	for email := range v.data {
		tmp = append(tmp, email)
	}
	return tmp
}

func ttl() int64 {
	return time.Now().Add(time.Hour).Unix()
}
