/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package schema

import (
	"database/sql"
	"time"

	"go.osspkg.com/goppy/sdk/errors"
)

var (
	ErrPoolNotFound = errors.New("pool not found")
)

const (
	MySQLDialect  = "mysql"
	SQLiteDialect = "sqlite"
	PgSQLDialect  = "pgsql"
)

type (
	//ConfigInterface interface of configs
	ConfigInterface interface {
		List() []ItemInterface
	}
	//ItemInterface config item interface
	ItemInterface interface {
		GetName() string
		GetDSN() string
		Setup(SetupInterface)
	}
	//SetupInterface connections setup interface
	SetupInterface interface {
		SetMaxIdleConns(int)
		SetMaxOpenConns(int)
		SetConnMaxLifetime(time.Duration)
	}
	//Connector interface of connection
	Connector interface {
		Dialect() string
		Pool(string) (*sql.DB, error)
		Reconnect() error
		Close() error
	}
)
