/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl

import (
	"context"
	"time"

	"go.osspkg.com/algorithms/structs/bitmap"
	"go.osspkg.com/ioutils/cache"
	"go.osspkg.com/syncing"
)

type object struct {
	store Storage
	opts  Options
	cache cache.Cache[string, *entity]
	mux   syncing.Funnel[string]
}

func New(ctx context.Context, store Storage, option Options) ACL {
	return &object{
		opts:  option,
		store: store,
		cache: cache.New[string, *entity](
			cache.OptTimeClean[string, *entity](ctx, option.CacheCheckInterval),
			cache.OptCountRandomClean[string, *entity](ctx, option.CacheSize, option.CacheCheckInterval),
		),
		mux: syncing.NewFunnel[string](),
	}
}

func (o *object) resolve(uid string) (*entity, error) {
	if ent, ok := o.cache.Get(uid); ok {
		return ent, nil
	}

	data, err := o.store.FindACL(uid)
	if err != nil {
		return nil, err
	}

	ent := &entity{
		bitmap: bitmap.New(),
		ts:     time.Now().Add(o.opts.CacheDeadline).Unix(),
	}

	if err = ent.bitmap.UnmarshalBinary(data); err != nil {
		return nil, err
	}

	o.cache.Set(uid, ent)

	return ent, nil
}

func (o *object) HasOne(uid string, feature uint16) (has bool, err error) {
	o.mux.Valve(uid, func() {
		var ent *entity
		if ent, err = o.resolve(uid); err != nil {
			return
		}

		has = ent.bitmap.Has(uint64(feature))
	})
	return
}

func (o *object) HasMany(uid string, features ...uint16) (has map[uint16]bool, err error) {
	o.mux.Valve(uid, func() {
		if len(features) < 1 {
			return
		}

		var ent *entity
		if ent, err = o.resolve(uid); err != nil {
			return
		}

		has = make(map[uint16]bool, len(features))
		for _, feature := range features {
			has[feature] = ent.bitmap.Has(uint64(feature))
		}
	})
	return
}

func (o *object) Set(uid string, features ...uint16) (err error) {
	o.mux.Valve(uid, func() {
		if len(features) < 1 {
			return
		}

		var ent *entity
		if ent, err = o.resolve(uid); err != nil {
			return
		}

		for _, feature := range features {
			ent.bitmap.Set(uint64(feature))
		}

		var b []byte
		if b, err = ent.bitmap.MarshalBinary(); err != nil {
			return
		}

		err = o.store.ChangeACL(uid, b)
	})
	return
}

func (o *object) Del(uid string, features ...uint16) (err error) {
	o.mux.Valve(uid, func() {
		if len(features) < 1 {
			return
		}

		var ent *entity
		if ent, err = o.resolve(uid); err != nil {
			return
		}

		for _, feature := range features {
			ent.bitmap.Del(uint64(feature))
		}

		var b []byte
		if b, err = ent.bitmap.MarshalBinary(); err != nil {
			return
		}

		err = o.store.ChangeACL(uid, b)
	})
	return
}
