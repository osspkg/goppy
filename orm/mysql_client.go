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

	_ "github.com/go-sql-driver/mysql" // nolint: golint
)

var (
	_ Connector       = (*_mysqlPool)(nil)
	_ ConfigInterface = (*ConfigMysqlClientPool)(nil)
)

type (
	// ConfigMysqlClientPool pool of configs
	ConfigMysqlClientPool struct {
		Pool []ConfigMysqlClient `yaml:"mysql"`
	}

	// ConfigMysqlClient config model
	ConfigMysqlClient struct {
		Tags              string        `yaml:"tags"`
		Host              string        `yaml:"host"`
		Port              int           `yaml:"port"`
		Schema            string        `yaml:"schema"`
		User              string        `yaml:"user"`
		Password          string        `yaml:"password"`
		Timezone          string        `yaml:"timezone"`
		TxIsolationLevel  string        `yaml:"txisolevel"`
		Charset           string        `yaml:"charset"`
		Collation         string        `yaml:"collation"`
		MaxIdleConn       int           `yaml:"maxidleconn"`
		MaxOpenConn       int           `yaml:"maxopenconn"`
		InterpolateParams bool          `yaml:"interpolateparams"`
		MaxConnTTL        time.Duration `yaml:"maxconnttl"`
		Timeout           time.Duration `yaml:"timeout"`
		ReadTimeout       time.Duration `yaml:"readtimeout"`
		WriteTimeout      time.Duration `yaml:"writetimeout"`
		OtherParams       string        `yaml:"other_params"`
	}

	_mysqlPool struct {
		configs map[string]ItemInterface
	}
)

func (c *ConfigMysqlClientPool) List() []ItemInterface {
	list := make([]ItemInterface, 0, len(c.Pool))
	for _, item := range c.Pool {
		list = append(list, item)
	}
	return list
}

func (i ConfigMysqlClient) GetTags() []string {
	return strings.Split(i.Tags, ",")
}

// Setup setting config conntections params
func (i ConfigMysqlClient) Setup(s SetupInterface) {
	s.SetMaxIdleConns(i.MaxIdleConn)
	s.SetMaxOpenConns(i.MaxOpenConn)
	s.SetConnMaxLifetime(i.MaxConnTTL)
}

// GetDSN connection params
func (i ConfigMysqlClient) GetDSN() string {
	params, err := url.ParseQuery(i.OtherParams)
	if err != nil {
		params = url.Values{}
	}

	params.Add("autocommit", "true")
	params.Add("interpolateParams", fmt.Sprintf("%t", i.InterpolateParams))

	// ---
	if len(i.Charset) == 0 {
		i.Charset = "utf8mb4"
	}
	params.Add("charset", i.Charset)
	// ---
	if len(i.Collation) == 0 {
		i.Collation = "utf8mb4_unicode_ci"
	}
	params.Add("collation", i.Collation)
	// ---
	if i.Timeout == 0 {
		i.Timeout = defaultTimeoutConn
	}
	params.Add("timeout", i.Timeout.String())
	// ---
	if i.ReadTimeout == 0 {
		i.ReadTimeout = defaultTimeout
	}
	params.Add("readTimeout", i.ReadTimeout.String())
	// ---
	if i.WriteTimeout == 0 {
		i.WriteTimeout = defaultTimeout
	}
	params.Add("writeTimeout", i.WriteTimeout.String())
	// ---
	if len(i.TxIsolationLevel) > 0 {
		params.Add("transaction_isolation", i.TxIsolationLevel)
	}
	// ---
	if len(i.Timezone) == 0 {
		i.Timezone = "UTC"
	}
	params.Add("loc", i.Timezone)
	// ---

	// ---
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", i.User, i.Password, i.Host, i.Port, i.Schema, params.Encode())
}

func NewMysqlClient(conf ConfigInterface) Connector {
	c := &_mysqlPool{
		configs: make(map[string]ItemInterface),
	}

	for _, item := range conf.List() {
		for _, tag := range item.GetTags() {
			c.configs[tag] = item
		}
	}

	return c
}

func (p *_mysqlPool) Dialect() string {
	return MySQLDialect
}

func (p *_mysqlPool) Tags() []string {
	tags := make([]string, 0, len(p.configs))
	for tag := range p.configs {
		tags = append(tags, tag)
	}
	return tags
}

func (p *_mysqlPool) Connect(_ context.Context, tag string) (*sql.DB, error) {
	conf, ok := p.configs[tag]
	if !ok {
		return nil, fmt.Errorf("tag not found")
	}
	db, err := sql.Open("mysql", conf.GetDSN())
	if err != nil {
		return nil, err
	}
	conf.Setup(db)
	return db, nil
}
