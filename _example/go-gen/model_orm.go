/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

// Code generated by goppy-cli for goppy.orm. DO NOT EDIT.
package go_gen

import (
	"context"
	time "time"

	uuid "github.com/google/uuid"

	"go.osspkg.com/goppy/v2/orm"
)

type RepositoryModel struct {
	orm        orm.ORM
	rtag, wtag string
}

func newRepositoryModel(orm orm.ORM) *RepositoryModel {
	return &RepositoryModel{
		orm:  orm,
		rtag: "slave",
		wtag: "master",
	}
}

func (v *RepositoryModel) TagSlave() orm.Stmt {
	return v.orm.Tag(v.rtag)
}

func (v *RepositoryModel) TagMaster() orm.Stmt {
	return v.orm.Tag(v.wtag)
}

const sqlUsersReadUserAll = `SELECT "name", "value"
			 FROM "users";
`

func (v *RepositoryModel) ReadUserAll(ctx context.Context) ([]User,
	error) {
	result := make([]User, 0, 2)
	err := v.orm.Tag(v.rtag).Query(ctx, "users_read_all", func(q orm.Querier) {
		q.SQL(sqlUsersReadUserAll)
		q.Bind(func(bind orm.Scanner) error {
			m := User{}
			if e := bind.Scan(&m.Name, &m.Value); e != nil {
				return e
			}
			result = append(result, m)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

const sqlUsersReadUserById = `SELECT "name", "value"
			 FROM "users"
			 WHERE "id" = ANY($1);
`

func (v *RepositoryModel) ReadUserById(
	ctx context.Context, args ...int64,
) ([]User, error) {
	result := make([]User, 0, 2)
	err := v.orm.Tag(v.rtag).Query(ctx, "users_read_by_id", func(q orm.Querier) {
		q.SQL(sqlUsersReadUserById, args)
		q.Bind(func(bind orm.Scanner) error {
			m := User{}
			if e := bind.Scan(&m.Name, &m.Value); e != nil {
				return e
			}
			result = append(result, m)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

const sqlUsersReadUserByName = `SELECT "name", "value"
			 FROM "users"
			 WHERE "name" = ANY($1);
`

func (v *RepositoryModel) ReadUserByName(
	ctx context.Context, args ...string,
) ([]User, error) {
	result := make([]User, 0, 2)
	err := v.orm.Tag(v.rtag).Query(ctx, "users_read_by_name", func(q orm.Querier) {
		q.SQL(sqlUsersReadUserByName, args)
		q.Bind(func(bind orm.Scanner) error {
			m := User{}
			if e := bind.Scan(&m.Name, &m.Value); e != nil {
				return e
			}
			result = append(result, m)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

const sqlUsersReadUserByValue = `SELECT "name", "value"
			 FROM "users"
			 WHERE "value" = ANY($1);
`

func (v *RepositoryModel) ReadUserByValue(
	ctx context.Context, args ...string,
) ([]User, error) {
	result := make([]User, 0, 2)
	err := v.orm.Tag(v.rtag).Query(ctx, "users_read_by_value", func(q orm.Querier) {
		q.SQL(sqlUsersReadUserByValue, args)
		q.Bind(func(bind orm.Scanner) error {
			m := User{}
			if e := bind.Scan(&m.Name, &m.Value); e != nil {
				return e
			}
			result = append(result, m)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

const sqlMetaCreateMeta = `INSERT INTO "meta" ("id", "user_id", "roles", "fail", "created_at", "updated_at", "deleted_at") 
			VALUES ($1, $2, $3, $4, $5, $6, $7) 
			RETURNING ("id");
`

func (v *RepositoryModel) CreateMeta(ctx context.Context, m *Meta) error {
	m.Id = uuid.New()
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()

	return v.orm.Tag(v.wtag).Query(ctx, "meta_create", func(q orm.Querier) {
		q.SQL(
			sqlMetaCreateMeta,
			m.Id, m.UserId, m.Roles, m.Fail, m.CreatedAt, m.UpdatedAt, m.DeletedAt,
		)
		q.Bind(func(bind orm.Scanner) error {
			return bind.Scan(&m.Id)
		})
	})
}

const sqlMetaUpdateMeta = `UPDATE "meta" SET 
			"id" = $1, "user_id" = $2, "roles" = $3, "fail" = $4, "created_at" = $5, "updated_at" = $6, "deleted_at" = $7
			 WHERE "id" = $8;
`

func (v *RepositoryModel) UpdateMeta(ctx context.Context, m *Meta) error {
	m.UpdatedAt = time.Now()

	return v.orm.Tag(v.wtag).Exec(ctx, "meta_update", func(e orm.Executor) {
		e.SQL(sqlMetaUpdateMeta)
		e.Params(m.Id, m.UserId, m.Roles, m.Fail, m.CreatedAt, m.UpdatedAt, m.DeletedAt, m.Id)
	})
}

const sqlMetaDeleteMeta = `DELETE FROM "meta"
			 WHERE "id" = $1;
`

func (v *RepositoryModel) DeleteMeta(ctx context.Context, pk uuid.UUID) error {
	return v.orm.Tag(v.wtag).Exec(ctx, "meta_delete", func(e orm.Executor) {
		e.SQL(sqlMetaDeleteMeta)
		e.Params(pk)
	})
}

const sqlMetaReadMetaAll = `SELECT "id", "user_id", "roles", "fail", "created_at", "updated_at", "deleted_at"
			 FROM "meta";
`

func (v *RepositoryModel) ReadMetaAll(ctx context.Context) ([]Meta,
	error) {
	result := make([]Meta, 0, 2)
	err := v.orm.Tag(v.rtag).Query(ctx, "meta_read_all", func(q orm.Querier) {
		q.SQL(sqlMetaReadMetaAll)
		q.Bind(func(bind orm.Scanner) error {
			m := Meta{}
			if e := bind.Scan(&m.Id, &m.UserId, &m.Roles, &m.Fail, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); e != nil {
				return e
			}
			result = append(result, m)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

const sqlMetaReadMetaById = `SELECT "id", "user_id", "roles", "fail", "created_at", "updated_at", "deleted_at"
			 FROM "meta"
			 WHERE "id" = ANY($1);
`

func (v *RepositoryModel) ReadMetaById(
	ctx context.Context, args ...uuid.UUID,
) ([]Meta, error) {
	result := make([]Meta, 0, 2)
	err := v.orm.Tag(v.rtag).Query(ctx, "meta_read_by_id", func(q orm.Querier) {
		q.SQL(sqlMetaReadMetaById, args)
		q.Bind(func(bind orm.Scanner) error {
			m := Meta{}
			if e := bind.Scan(&m.Id, &m.UserId, &m.Roles, &m.Fail, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); e != nil {
				return e
			}
			result = append(result, m)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

const sqlMetaReadMetaByUserId = `SELECT "id", "user_id", "roles", "fail", "created_at", "updated_at", "deleted_at"
			 FROM "meta"
			 WHERE "user_id" = ANY($1);
`

func (v *RepositoryModel) ReadMetaByUserId(
	ctx context.Context, args ...int64,
) ([]Meta, error) {
	result := make([]Meta, 0, 2)
	err := v.orm.Tag(v.rtag).Query(ctx, "meta_read_by_user_id", func(q orm.Querier) {
		q.SQL(sqlMetaReadMetaByUserId, args)
		q.Bind(func(bind orm.Scanner) error {
			m := Meta{}
			if e := bind.Scan(&m.Id, &m.UserId, &m.Roles, &m.Fail, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); e != nil {
				return e
			}
			result = append(result, m)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

const sqlMetaReadMetaByRoles = `SELECT "id", "user_id", "roles", "fail", "created_at", "updated_at", "deleted_at"
			 FROM "meta"
			 WHERE "roles" = ANY($1);
`

func (v *RepositoryModel) ReadMetaByRoles(
	ctx context.Context, args ...string,
) ([]Meta, error) {
	result := make([]Meta, 0, 2)
	err := v.orm.Tag(v.rtag).Query(ctx, "meta_read_by_roles", func(q orm.Querier) {
		q.SQL(sqlMetaReadMetaByRoles, args)
		q.Bind(func(bind orm.Scanner) error {
			m := Meta{}
			if e := bind.Scan(&m.Id, &m.UserId, &m.Roles, &m.Fail, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); e != nil {
				return e
			}
			result = append(result, m)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

const sqlMetaReadMetaByFail = `SELECT "id", "user_id", "roles", "fail", "created_at", "updated_at", "deleted_at"
			 FROM "meta"
			 WHERE "fail" = ANY($1);
`

func (v *RepositoryModel) ReadMetaByFail(
	ctx context.Context, args ...bool,
) ([]Meta, error) {
	result := make([]Meta, 0, 2)
	err := v.orm.Tag(v.rtag).Query(ctx, "meta_read_by_fail", func(q orm.Querier) {
		q.SQL(sqlMetaReadMetaByFail, args)
		q.Bind(func(bind orm.Scanner) error {
			m := Meta{}
			if e := bind.Scan(&m.Id, &m.UserId, &m.Roles, &m.Fail, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); e != nil {
				return e
			}
			result = append(result, m)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
