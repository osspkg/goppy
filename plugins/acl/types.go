/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl

import (
	"time"

	"go.osspkg.com/algorithms/structs/bitmap"
)

type ACL interface {
	HasOne(uid string, feature uint16) (bool, error)
	HasMany(uid string, features ...uint16) (has map[uint16]bool, err error)
	Set(uid string, features ...uint16) error
	Del(uid string, features ...uint16) error
}

type Storage interface {
	FindACL(uid string) ([]byte, error)
	ChangeACL(uid string, access []byte) error
}

type entity struct {
	bitmap *bitmap.Bitmap
	ts     int64
}

func (e entity) Timestamp() int64 {
	return e.ts
}

type Options struct {
	CacheSize          int
	CacheDeadline      time.Duration
	CacheCheckInterval time.Duration
}
