/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package pgsql

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.osspkg.com/goppy/v3/orm/dialect"
)

type ConfigGroup struct {
	Pool []Config `yaml:"pgsql"`
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
				Tags:        "master",
				Host:        "127.0.0.1",
				Port:        5432,
				Schema:      "postgres",
				User:        "postgres",
				Password:    "postgres",
				SSLMode:     false,
				AppName:     "goppy_app",
				MaxIdleConn: 5,
				MaxOpenConn: 5,
				MaxConnTTL:  time.Second * 50,
				Charset:     "UTF8",
				Timeout:     time.Second * 5,
				OtherParams: "",
			},
		}
	}
}

type Config struct {
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
		i.Timeout = dialect.DefaultTimeoutConn
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
