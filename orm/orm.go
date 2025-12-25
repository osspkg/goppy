/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"fmt"
	"time"

	"go.osspkg.com/do"
	"go.osspkg.com/errors"
	"go.osspkg.com/logx"
	"go.osspkg.com/routine/tick"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v2/orm/dialect"
)

var (
	ErrTagNotFound = errors.New("tag not found")
)

type (
	_orm struct {
		pool  *syncing.Map[string, Stmt]
		conns *syncing.Map[string, dialect.Connector]
		ctx   context.Context
	}

	ORM interface {
		Tag(name string) Stmt
		ApplyConfig(dialectName dialect.Name, c dialect.ConfigInterface) error
		Close()
	}
)

// New init database connections
func New(ctx context.Context) ORM {
	db := &_orm{
		pool:  syncing.NewMap[string, Stmt](10),
		conns: syncing.NewMap[string, dialect.Connector](10),
		ctx:   ctx,
	}

	do.Async(func() {
		tik := tick.Ticker{
			Calls: []tick.Config{
				{
					Name:     "validate DB connects",
					Interval: time.Second * 15,
					Func: func(ctx context.Context, _ time.Time) error {
						db.checkConnects(ctx)
						return nil
					},
				},
			},
		}
		tik.Run(ctx)
	}, func(err error) {
		logx.Error("Validate DB connects", "err", err)
	})

	return db
}

func (v *_orm) Close() {
	v.conns.Reset()

	for _, tag := range v.pool.Keys() {
		stmt, ok := v.pool.Extract(tag)
		if !ok {
			continue
		}

		if err := stmt.Close(); err != nil {
			logx.Error("Close DB connect", "err", err, "tag", tag)
		}
	}
}

// Tag getting stmt by name
func (v *_orm) Tag(name string) Stmt {
	if s, ok := v.pool.Get(name); ok {
		return s
	}
	return newStmt(name, nil, nil, ErrTagNotFound)
}

func (v *_orm) ApplyConfig(dialectName dialect.Name, cfg dialect.ConfigInterface) error {
	c, ok := dialect.GetConnector(dialectName)
	if !ok {
		return fmt.Errorf("dialect '%s' not registered", dialectName)
	}

	c.ApplyConfig(cfg)

	for _, tag := range c.Tags() {
		cc, has := v.pool.Get(tag)

		stmt, err := v.initConnect(c, tag)
		if err != nil {
			return fmt.Errorf("create db connect [%s:%s]: %w", c.Dialect(), tag, err)
		}

		v.pool.Set(tag, stmt)
		v.conns.Set(tag, c)

		logx.Info("Create new DB connect", "dialect", c.Dialect(), "tag", tag)

		if has {
			if err = cc.Close(); err != nil {
				logx.Error("Close old DB connect", "dialect", c.Dialect(), "tag", tag, "err", err)
			}
		}

	}

	return nil
}

func (v *_orm) initConnect(c dialect.Connector, tag string) (Stmt, error) {
	db, err := c.Connect(v.ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}

	if err = db.PingContext(v.ctx); err != nil {
		return nil, fmt.Errorf("connect ping failed: %w", err)
	}

	return newStmt(tag, c, db, nil), nil
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

		stmt, err := v.initConnect(c, tag)
		if err != nil {
			logx.Error("Create DB connect", "err", err, "tag", tag)
			continue
		}

		v.pool.Set(tag, stmt)
	}
}
