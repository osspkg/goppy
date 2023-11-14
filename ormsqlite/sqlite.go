/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ormsqlite

import (
	"database/sql"
	"fmt"
	"net/url"
	"sync"

	_ "github.com/mattn/go-sqlite3" //nolint: golint
	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/sqlcommon"
)

var (
	_ sqlcommon.Connector       = (*pool)(nil)
	_ sqlcommon.ConfigInterface = (*Config)(nil)
)

type (
	//Config pool of configs
	Config struct {
		Pool []Item `yaml:"sqlite"`
	}

	//Item config model
	Item struct {
		Name        string `yaml:"name"`
		File        string `yaml:"file"`
		Cache       string `yaml:"cache"`
		Mode        string `yaml:"mode"`
		Journal     string `yaml:"journal"`
		LockingMode string `yaml:"locking_mode"`
		OtherParams string `yaml:"other_params"`
	}

	pool struct {
		conf sqlcommon.ConfigInterface
		db   map[string]*sql.DB
		l    sync.RWMutex
	}
)

// List getting all configs
func (c *Config) List() (list []sqlcommon.ItemInterface) {
	for _, item := range c.Pool {
		list = append(list, item)
	}
	return
}

// GetName getting config name
func (i Item) GetName() string { return i.Name }

// GetDSN connection params
func (i Item) GetDSN() string {
	params, err := url.ParseQuery(i.OtherParams)
	if err != nil {
		params = url.Values{}
	}
	//---
	if len(i.Cache) == 0 {
		i.Cache = "private"
	}
	params.Add("cache", i.Cache)
	//---
	if len(i.Mode) == 0 {
		i.Mode = "rwc"
	}
	params.Add("mode", i.Mode)
	//---
	if len(i.Journal) == 0 {
		i.Journal = "TRUNCATE"
	}
	params.Add("_journal", i.Journal)
	//---
	if len(i.LockingMode) == 0 {
		i.LockingMode = "EXCLUSIVE"
	}
	params.Add("_locking_mode", i.LockingMode)
	//--
	return fmt.Sprintf("file:%s?%s", i.File, params.Encode())
}

// Setup setting config conntections params
func (i Item) Setup(_ sqlcommon.SetupInterface) {}

// New init new sqlite connection
func New(conf sqlcommon.ConfigInterface) sqlcommon.Connector {
	c := &pool{
		conf: conf,
		db:   make(map[string]*sql.DB),
	}

	return c
}

// Dialect getting sql dialect
func (p *pool) Dialect() string {
	return sqlcommon.SQLiteDialect
}

// Reconnect update connection to database
func (p *pool) Reconnect() error {
	if err := p.Close(); err != nil {
		return err
	}

	p.l.Lock()
	defer p.l.Unlock()

	for _, item := range p.conf.List() {
		db, err := sql.Open("sqlite3", item.GetDSN())
		if err != nil {
			if er := p.Close(); er != nil {
				return errors.Wrap(err, er)
			}
			return err
		}
		p.db[item.GetName()] = db
	}
	return nil
}

// Close closing connection
func (p *pool) Close() error {
	p.l.Lock()
	defer p.l.Unlock()

	if len(p.db) > 0 {
		for _, db := range p.db {
			if err := db.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Pool getting connection pool by name
func (p *pool) Pool(name string) (*sql.DB, error) {
	p.l.RLock()
	defer p.l.RUnlock()

	db, ok := p.db[name]
	if !ok {
		return nil, sqlcommon.ErrPoolNotFound
	}
	return db, db.Ping()
}
