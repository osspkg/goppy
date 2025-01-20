/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"fmt"
	"time"

	"go.osspkg.com/errors"
	"go.osspkg.com/logx"
	"go.osspkg.com/routine"
	"go.osspkg.com/syncing"
)

var (
	ErrTagNotFound = errors.New("tag not found")
)

type (
	_orm struct {
		pool  *syncing.Map[string, Stmt]
		conns *syncing.Map[string, Connector]
		ctx   context.Context
	}

	ORM interface {
		Tag(name string) Stmt
		Register(c Connector, onSuccess func()) error
		Close()
	}
)

// New init database connections
func New(ctx context.Context) ORM {
	db := &_orm{
		pool:  syncing.NewMap[string, Stmt](10),
		conns: syncing.NewMap[string, Connector](10),
		ctx:   ctx,
	}

	routine.Interval(ctx, time.Second*15, db.checkConnects)

	return db
}

func (v *_orm) Close() {
	for _, tag := range v.pool.Keys() {
		stmt, ok := v.pool.Extract(tag)
		if !ok {
			continue
		}

		if err := stmt.Close(); err != nil {
			logx.Error("Close DB connect", "err", err, "tag", tag)
		}

		v.conns.Del(tag)
	}
}

// Tag getting stmt by name
func (v *_orm) Tag(name string) Stmt {
	if s, ok := v.pool.Get(name); ok {
		return s
	}
	return newStmt(name, "", nil, ErrTagNotFound)
}

func (v *_orm) Register(c Connector, onSuccess func()) error {
	for _, tag := range c.Tags() {
		if _, ok := v.conns.Get(tag); ok {
			return fmt.Errorf("db connect alredy exist: %s", tag)
		}

		if err := v.appendConnect(c, tag); err != nil {
			return fmt.Errorf("create db connect [%s:%s]: %w", c.Dialect(), tag, err)
		}

		v.conns.Set(tag, c)
		logx.Info("Create DB connect", "dialect", c.Dialect(), "tag", tag)
	}

	go onSuccess()

	return nil
}

func (v *_orm) appendConnect(c Connector, tag string) error {
	db, err := c.Connect(v.ctx, tag)
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	if err = db.PingContext(v.ctx); err != nil {
		return fmt.Errorf("connect ping failed: %w", err)
	}

	v.pool.Set(tag, newStmt(tag, c.Dialect(), db, nil))
	return nil
}

func (v *_orm) checkConnects(ctx context.Context) {
	badConns := make(map[string]struct{}, 10)

	for _, tag := range v.pool.Keys() {
		stmt, ok := v.pool.Get(tag)
		if !ok {
			badConns[tag] = struct{}{}
			continue
		}

		if err := stmt.PingContext(ctx); err != nil {
			logx.Error("Bad DB connect", "err", err, "tag", tag)
			badConns[tag] = struct{}{}
		}
	}

	if len(badConns) == 0 {
		return
	}

	for tag := range badConns {
		c, ok := v.conns.Get(tag)
		if !ok {
			continue
		}

		if err := v.appendConnect(c, tag); err != nil {
			logx.Error("Create DB connect", "err", err, "tag", tag)
		}
	}
}
