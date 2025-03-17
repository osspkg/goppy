/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package go_gen

import (
	"time"

	"github.com/google/uuid"
)

//go:generate goppy gen --type=orm:pgsql --db-read=slave --db-write=master --index=1000

//gen:orm table=users action=ro
type User struct {
	Id    int64   // col=id index=pk
	Name  string  // col=name len=100
	Meta0 []*Meta // link=id:meta.user_id
	//Meta1 []Meta  // link=id
	//Meta2 *[]Meta // link=id
	//Meta3 Meta    // link=id
	//Meta4 *Meta   // link=id
}

//gen:orm table=meta index=uniq:id,user_id index=uniq:user_id,id
type Meta struct {
	Id        uuid.UUID  // col=id index=pk auto=c:uuid.New()
	UserId    int64      // col=user_id index=fk:users.id
	Roles     []string   // col=roles index=uniq
	Fail      bool       // col=fail
	CreatedAt time.Time  // col=created_at auto=c:time.Now()
	UpdatedAt time.Time  // col=updated_at auto=time.Now()
	DeletedAt *time.Time // col=deleted_at
}
