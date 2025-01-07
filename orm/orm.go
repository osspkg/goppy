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
		pool  map[string]Stmt
		conns map[string]Connector
		opts  *options
		mux   syncing.Lock
		ctx   context.Context
	}

	ORM interface {
		Tag(name string) Stmt
		Register(c Connector)
		Close()
	}

	options struct {
		Metrics metricExecutor
	}

	Option func(o *options)
)

func UseMetric(name string) Option {
	return func(o *options) {
		o.Metrics = newMetric(name)
	}
}

// New init database connections
func New(ctx context.Context, opts ...Option) ORM {
	o := &options{
		Metrics: DevNullMetric,
	}

	for _, opt := range opts {
		opt(o)
	}

	db := &_orm{
		pool:  make(map[string]Stmt, 10),
		conns: make(map[string]Connector, 10),
		opts:  o,
		mux:   syncing.NewLock(),
		ctx:   ctx,
	}

	routine.Interval(ctx, time.Second*15, db.checkConnects)

	return db
}

func (v *_orm) Close() {
	v.mux.Lock(func() {
		for tag, stmt := range v.pool {
			if err := stmt.Close(); err != nil {
				logx.Error("Close DB connect", "err", err, "tag", tag)
			}
			delete(v.pool, tag)
			delete(v.conns, tag)
		}
	})
}

// Tag getting stmt by name
func (v *_orm) Tag(name string) (s Stmt) {
	v.mux.RLock(func() {
		var ok bool
		s, ok = v.pool[name]
		if !ok {
			s = newStmt("", nil, v.opts, ErrTagNotFound)
		}
	})
	return
}

func (v *_orm) Register(c Connector) {
	for _, tag := range c.Tags() {
		if err := v.appendConnect(c, tag); err != nil {
			logx.Error("Create DB connect", "err", err, "tag", tag)
			continue
		}
		v.mux.Lock(func() {
			v.conns[tag] = c
		})
		logx.Info("Create DB connect", "dialect", c.Dialect(), "tag", tag)
	}
	return
}

func (v *_orm) appendConnect(c Connector, tag string) error {
	db, err := c.Connect(v.ctx, tag)
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	if err = db.PingContext(v.ctx); err != nil {
		return fmt.Errorf("connect ping failed: %w", err)
	}
	v.mux.RLock(func() {
		if _, ok := v.pool[tag]; ok {
			err = fmt.Errorf("pool exist")
		}
	})
	stmt := newStmt(c.Dialect(), db, v.opts, nil)
	v.mux.Lock(func() {
		v.pool[tag] = stmt
	})
	return nil
}

func (v *_orm) checkConnects(ctx context.Context) {
	var badTags map[string]Connector
	v.mux.RLock(func() {
		badTags = make(map[string]Connector, len(v.conns))
		for tag, st := range v.pool {
			if err := st.PingContext(ctx); err != nil {
				logx.Error("Bad DB connect", "err", err, "tag", tag)
				badTags[tag] = nil
			}
		}
	})
	if len(badTags) == 0 {
		return
	}
	v.mux.Lock(func() {
		for tag := range badTags {
			delete(v.pool, tag)
			c, ok := v.conns[tag]
			if !ok {
				delete(badTags, tag)
				continue
			}
			badTags[tag] = c
		}
	})
	if len(badTags) == 0 {
		return
	}
	for tag, c := range badTags {
		if err := v.appendConnect(c, tag); err != nil {
			logx.Error("Create DB connect", "err", err, "tag", tag)
		}
	}
}
