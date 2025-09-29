/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package pgsql

type migrate struct {
}

func (migrate) CreateTableQuery() []string {
	return []string{
		`CREATE SEQUENCE IF NOT EXISTS "__migrations___id_seq" INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;`,
		`CREATE TABLE "__migrations__" (
			"id" integer DEFAULT nextval('__migrations___id_seq') NOT NULL,
			"name" text NOT NULL,
			"timestamp" integer NOT NULL,
			CONSTRAINT "__migrations___pkey" PRIMARY KEY ("id")
		) WITH (oids = false);`,
	}
}

func (migrate) CheckTableQuery() string {
	return `SELECT "tablename" FROM "pg_catalog"."pg_tables" WHERE tablename='__migrations__';`
}

func (migrate) CompletedQuery() string {
	return `SELECT "name" FROM "__migrations__";`
}

func (migrate) SaveQuery() string {
	return `INSERT INTO "__migrations__" ("name", "timestamp") VALUES ($1, $2);`
}
