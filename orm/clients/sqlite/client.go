/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // nolint: golint
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v3/orm/dialect"
)

type pool struct {
	configs *syncing.Map[string, dialect.ItemInterface]
}

func New() dialect.Connector {
	c := &pool{
		configs: syncing.NewMap[string, dialect.ItemInterface](0),
	}

	return c
}

func (p *pool) EmptyConfig() dialect.ConfigInterface {
	return &ConfigGroup{}
}

func (p *pool) ApplyConfig(cfg dialect.ConfigInterface) {
	for _, item := range cfg.List() {
		for _, tag := range item.GetTags() {
			p.configs.Set(tag, item)
		}
	}
}

func (p *pool) Dialect() dialect.Name {
	return Name
}

func (p *pool) Tags() []string {
	return p.configs.Keys()
}

func (p *pool) Connect(_ context.Context, tag string) (*sql.DB, error) {
	conf, ok := p.configs.Get(tag)
	if !ok {
		return nil, fmt.Errorf("tag not found")
	}
	db, err := sql.Open("sqlite3", conf.GetDSN())
	if err != nil {
		return nil, err
	}
	conf.Setup(db)
	return db, nil
}

func (p *pool) CastTypesFunc() func(args []any) {
	return nil
}

func (p *pool) HasLastInsertId() bool {
	return true
}
