/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl

import (
	"context"
	"time"

	"go.osspkg.com/goppy/errors"
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

type object struct {
	cache *cache
	store Storage
}

func New(store Storage, size uint) ACL {
	return &object{
		store: store,
		cache: newCache(size),
	}
}

func (v *object) AutoFlush(ctx context.Context, interval time.Duration) {
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

func (v *object) GetAll(email string) ([]uint8, error) {
	if !v.cache.Has(email) {
		if err := v.loadFromStore(email); err != nil {
			return nil, err
		}
	}

	return v.cache.GetAll(email)
}

func (v *object) Get(email string, feature uint16) (uint8, error) {
	if !v.cache.Has(email) {
		if err := v.loadFromStore(email); err != nil {
			return 0, err
		}
	}

	return v.cache.Get(email, feature)
}

func (v *object) Set(email string, feature uint16, level uint8) error {
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

func (v *object) Flush(email string) {
	v.cache.Flush(email)
}

func (v *object) loadFromStore(email string) error {
	access, err := v.store.FindACL(email)
	if err != nil {
		return errors.Wrap(err, errUserNotFound)
	}
	v.cache.SetAll(email, str2uint(access)...)
	return nil
}

func (v *object) saveToStore(email string) error {
	access, err := v.cache.GetAll(email)
	if err != nil {
		return err
	}

	err = v.store.ChangeACL(email, uint2str(access...))
	return errors.Wrapf(err, "change acl")
}
