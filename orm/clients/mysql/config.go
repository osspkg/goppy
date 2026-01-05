/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package mysql

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.osspkg.com/goppy/v3/orm/dialect"
)

type ConfigGroup struct {
	Pool []Config `yaml:"mysql"`
}

func (c *ConfigGroup) List() []dialect.ItemInterface {
	list := make([]dialect.ItemInterface, 0, len(c.Pool))
	for _, item := range c.Pool {
		list = append(list, item)
	}
	return list
}

func (c *ConfigGroup) Default() {
	if len(c.Pool) == 0 {
		c.Pool = []Config{
			{
				Tags:              "master",
				Host:              "127.0.0.1",
				Port:              3306,
				Schema:            "test_database",
				User:              "test",
				Password:          "test",
				MaxIdleConn:       5,
				MaxOpenConn:       5,
				MaxConnTTL:        time.Second * 50,
				InterpolateParams: false,
				Timezone:          "UTC",
				TxIsolationLevel:  "",
				Charset:           "utf8mb4",
				Collation:         "utf8mb4_unicode_ci",
				Timeout:           time.Second * 5,
				ReadTimeout:       time.Second * 5,
				WriteTimeout:      time.Second * 5,
				OtherParams:       "",
			},
		}
	}
}

type Config struct {
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

func (i Config) GetTags() []string {
	return strings.Split(i.Tags, ",")
}

// Setup setting config conntections params
func (i Config) Setup(s dialect.SetupInterface) {
	s.SetMaxIdleConns(i.MaxIdleConn)
	s.SetMaxOpenConns(i.MaxOpenConn)
	s.SetConnMaxLifetime(i.MaxConnTTL)
}

// GetDSN connection params
func (i Config) GetDSN() string {
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
		i.Timeout = dialect.DefaultTimeoutConn
	}
	params.Add("timeout", i.Timeout.String())
	// ---
	if i.ReadTimeout == 0 {
		i.ReadTimeout = dialect.DefaultTimeout
	}
	params.Add("readTimeout", i.ReadTimeout.String())
	// ---
	if i.WriteTimeout == 0 {
		i.WriteTimeout = dialect.DefaultTimeout
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
