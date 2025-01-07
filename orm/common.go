/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"database/sql"
	"time"
)

const (
	defaultTimeout     = time.Second * 5
	defaultTimeoutConn = time.Second * 60
)

const (
	MySQLDialect  = "mysql"
	SQLiteDialect = "sqlite"
	PgSQLDialect  = "pgsql"
)

type (
	// ConfigInterface interface of configs
	ConfigInterface interface {
		List() []ItemInterface
	}
	// ItemInterface config item interface
	ItemInterface interface {
		GetTags() []string
		GetDSN() string
		Setup(SetupInterface)
	}
	// SetupInterface connections setup interface
	SetupInterface interface {
		SetMaxIdleConns(int)
		SetMaxOpenConns(int)
		SetConnMaxLifetime(time.Duration)
	}
	// Connector interface of connection
	Connector interface {
		Dialect() string
		Connect(ctx context.Context, tag string) (*sql.DB, error)
		Tags() []string
	}
	// Migrator interface of migration
	Migrator interface {
		CreateTableQuery() []string
		CheckTableQuery() string
		CompletedQuery() string
		SaveQuery() string
	}
)
