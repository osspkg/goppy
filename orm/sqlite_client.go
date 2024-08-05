/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/mattn/go-sqlite3" // nolint: golint
)

var (
	_ Connector       = (*_sqlitePool)(nil)
	_ ConfigInterface = (*ConfigSqliteClientPool)(nil)
)

type (
	// ConfigSqliteClientPool pool of configs
	ConfigSqliteClientPool struct {
		Pool []ConfigSqliteClient `yaml:"sqlite"`
	}

	// ConfigSqliteClient config model
	ConfigSqliteClient struct {
		Tags        string `yaml:"tags"`
		File        string `yaml:"file"`
		Cache       string `yaml:"cache"`
		Mode        string `yaml:"mode"`
		Journal     string `yaml:"journal"`
		LockingMode string `yaml:"locking_mode"`
		OtherParams string `yaml:"other_params"`
	}

	_sqlitePool struct {
		configs map[string]ItemInterface
	}
)

func (c *ConfigSqliteClientPool) List() []ItemInterface {
	list := make([]ItemInterface, 0, len(c.Pool))
	for _, item := range c.Pool {
		list = append(list, item)
	}
	return list
}

func (i ConfigSqliteClient) GetTags() []string {
	return strings.Split(i.Tags, ",")
}

func (i ConfigSqliteClient) GetDSN() string {
	params, err := url.ParseQuery(i.OtherParams)
	if err != nil {
		params = url.Values{}
	}
	// ---
	if len(i.Cache) == 0 {
		i.Cache = "private"
	}
	params.Add("cache", i.Cache)
	// ---
	if len(i.Mode) == 0 {
		i.Mode = "rwc"
	}
	params.Add("mode", i.Mode)
	// ---
	if len(i.Journal) == 0 {
		i.Journal = "TRUNCATE"
	}
	params.Add("_journal", i.Journal)
	// ---
	if len(i.LockingMode) == 0 {
		i.LockingMode = "EXCLUSIVE"
	}
	params.Add("_locking_mode", i.LockingMode)
	// --
	return fmt.Sprintf("file:%s?%s", i.File, params.Encode())
}

func (i ConfigSqliteClient) Setup(_ SetupInterface) {}

func NewSqliteClient(conf ConfigInterface) Connector {
	c := &_sqlitePool{
		configs: make(map[string]ItemInterface),
	}

	for _, item := range conf.List() {
		for _, tag := range item.GetTags() {
			c.configs[tag] = item
		}
	}

	return c
}

// Dialect getting sql dialect
func (p *_sqlitePool) Dialect() string {
	return SQLiteDialect
}

func (p *_sqlitePool) Tags() []string {
	tags := make([]string, 0, len(p.configs))
	for tag := range p.configs {
		tags = append(tags, tag)
	}
	return tags
}

func (p *_sqlitePool) Connect(_ context.Context, tag string) (*sql.DB, error) {
	conf, ok := p.configs[tag]
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
