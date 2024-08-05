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
	"time"

	_ "github.com/lib/pq" // nolint: golint
)

var (
	_ Connector       = (*_pgsqlPool)(nil)
	_ ConfigInterface = (*ConfigPGSqlClientPool)(nil)
)

type (
	// ConfigPGSqlClientPool pool of configs
	ConfigPGSqlClientPool struct {
		Pool []ConfigPGSqlClient `yaml:"postgresql"`
	}

	// ConfigPGSqlClient config model
	ConfigPGSqlClient struct {
		Tags        string        `yaml:"tags"`
		Host        string        `yaml:"host"`
		Port        int           `yaml:"port"`
		Schema      string        `yaml:"schema"`
		User        string        `yaml:"user"`
		Password    string        `yaml:"password"`
		SSLMode     bool          `yaml:"sslmode"`
		AppName     string        `yaml:"app_name"`
		Charset     string        `yaml:"charset"`
		MaxIdleConn int           `yaml:"maxidleconn"`
		MaxOpenConn int           `yaml:"maxopenconn"`
		MaxConnTTL  time.Duration `yaml:"maxconnttl"`
		Timeout     time.Duration `yaml:"timeout"`
		OtherParams string        `yaml:"other_params"`
	}

	_pgsqlPool struct {
		configs map[string]ItemInterface
	}
)

func (c *ConfigPGSqlClientPool) List() []ItemInterface {
	list := make([]ItemInterface, 0, len(c.Pool))
	for _, item := range c.Pool {
		list = append(list, item)
	}
	return list
}

func (i ConfigPGSqlClient) GetTags() []string {
	return strings.Split(i.Tags, ",")
}

// Setup setting config conntections params
func (i ConfigPGSqlClient) Setup(s SetupInterface) {
	s.SetMaxIdleConns(i.MaxIdleConn)
	s.SetMaxOpenConns(i.MaxOpenConn)
	s.SetConnMaxLifetime(i.MaxConnTTL)
}

// GetDSN connection params
func (i ConfigPGSqlClient) GetDSN() string {
	params, err := url.ParseQuery(i.OtherParams)
	if err != nil {
		params = url.Values{}
	}

	// ---
	if len(i.Charset) == 0 {
		i.Charset = "UTF8"
	}
	params.Add("client_encoding", i.Charset)
	// ---
	if i.SSLMode {
		params.Add("sslmode", "prefer")
	} else {
		params.Add("sslmode", "disable")
	}
	// ---
	if i.Timeout == 0 {
		i.Timeout = defaultTimeoutConn
	}
	params.Add("connect_timeout", fmt.Sprintf("%.0f", i.Timeout.Seconds()))
	// ---
	if len(i.AppName) == 0 {
		i.AppName = "go_app"
	}
	params.Add("application_name", i.AppName)
	// ---

	// ---
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s", i.User, i.Password, i.Host, i.Port, i.Schema, params.Encode())
}

func NewPGSqlClient(conf ConfigInterface) Connector {
	c := &_pgsqlPool{
		configs: make(map[string]ItemInterface),
	}

	for _, item := range conf.List() {
		for _, tag := range item.GetTags() {
			c.configs[tag] = item
		}
	}

	return c
}

func (p *_pgsqlPool) Dialect() string {
	return PgSQLDialect
}

func (p *_pgsqlPool) Tags() []string {
	tags := make([]string, 0, len(p.configs))
	for tag := range p.configs {
		tags = append(tags, tag)
	}
	return tags
}

func (p *_pgsqlPool) Connect(_ context.Context, tag string) (*sql.DB, error) {
	conf, ok := p.configs[tag]
	if !ok {
		return nil, fmt.Errorf("tag not found")
	}
	db, err := sql.Open("postgres", conf.GetDSN())
	if err != nil {
		return nil, err
	}
	conf.Setup(db)
	return db, nil
}
