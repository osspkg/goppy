/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl

import (
	"context"
	"time"

	"github.com/osspkg/goppy/sdk/errors"
)

var (
	errFeatureGreaterMax  = errors.New("feature number is greater than the maximum")
	errUserNotFound       = errors.New("user not found")
	errChangeNotSupported = errors.New("changing ACL is not supported")
)

type (
	ACL interface {
		GetAll(email string) ([]uint8, error)
		Get(email string, feature uint16) (uint8, error)
		Set(email string, feature uint16, level uint8) error
		Flush(email string)
		AutoFlush(ctx context.Context, interval time.Duration)
	}

	Storage interface {
		FindACL(email string) (string, error)
		ChangeACL(email, access string) error
	}
)

type _acl struct {
	cache *cache
	store Storage
}

func NewACL(store Storage, size uint) ACL {
	return &_acl{
		store: store,
		cache: newCache(size),
	}
}

func (v *_acl) AutoFlush(ctx context.Context, interval time.Duration) {
	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ts := <-tick.C:
			v.cache.FlushByTime(ts.Unix())
		}
	}
}

func (v *_acl) GetAll(email string) ([]uint8, error) {
	if !v.cache.Has(email) {
		if err := v.loadFromStore(email); err != nil {
			return nil, err
		}
	}

	return v.cache.GetAll(email)
}

func (v *_acl) Get(email string, feature uint16) (uint8, error) {
	if !v.cache.Has(email) {
		if err := v.loadFromStore(email); err != nil {
			return 0, err
		}
	}

	return v.cache.Get(email, feature)
}

func (v *_acl) Set(email string, feature uint16, level uint8) error {
	if !v.cache.Has(email) {
		if err := v.loadFromStore(email); err != nil {
			return err
		}
	}

	if err := v.cache.Set(email, feature, level); err != nil {
		return err
	}
	return v.saveToStore(email)
}

func (v *_acl) Flush(email string) {
	v.cache.Flush(email)
}

func (v *_acl) loadFromStore(email string) error {
	access, err := v.store.FindACL(email)
	if err != nil {
		return errors.Wrap(err, errUserNotFound)
	}
	v.cache.SetAll(email, str2uint(access)...)
	return nil
}

func (v *_acl) saveToStore(email string) error {
	access, err := v.cache.GetAll(email)
	if err != nil {
		return err
	}

	err = v.store.ChangeACL(email, uint2str(access...))
	return errors.Wrapf(err, "change acl")
}
