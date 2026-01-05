/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package go_gen

import (
	"time"

	"github.com/google/uuid"
)

//go:generate goppy gen-orm --dialect=pgsql --db-read=slave --db-write=master --index=1000

//gen:orm table=users crud=crud
type User struct {
	Id    int64    // col=id index=pk
	Name  string   // col=name len=100
	Value []string // col=value
	Meta0 []*Meta  // col=meta0
}

//gen:orm table=meta index=unq:id,user_id index=unq:user_id,id
type Meta struct {
	Id        int64      // col=id index=pk
	UID       uuid.UUID  // col=uid index=idx auto=c:uuid.New()
	UserId    int64      // col=user_id index=fk:users.id
	Roles     []string   // col=roles index=unq
	Fail      bool       // col=fail
	CreatedAt time.Time  // col=created_at auto=c:time.Now()
	UpdatedAt time.Time  // col=updated_at auto=time.Now()
	DeletedAt *time.Time // col=deleted_at
}
