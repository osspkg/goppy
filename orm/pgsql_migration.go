/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

type _pgsqlMigrateTable struct {
}

func (*_pgsqlMigrateTable) CreateTableQuery() []string {
	return []string{
		`CREATE SEQUENCE __migrations___id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;`,
		`CREATE TABLE "__migrations__" (
			"id" integer DEFAULT nextval('__migrations___id_seq') NOT NULL,
			"name" text NOT NULL,
			"timestamp" integer NOT NULL,
			CONSTRAINT "__migrations___pkey" PRIMARY KEY ("id")
		) WITH (oids = false);`,
	}
}

func (*_pgsqlMigrateTable) CheckTableQuery() string {
	return `SELECT "tablename" FROM "pg_catalog"."pg_tables" WHERE tablename='__migrations__';`
}

func (*_pgsqlMigrateTable) CompletedQuery() string {
	return `SELECT "name" FROM "__migrations__";`
}

func (*_pgsqlMigrateTable) SaveQuery() string {
	return `INSERT INTO "__migrations__" ("name", "timestamp") VALUES ($1, $2);`
}
