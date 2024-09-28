/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

type _sqliteMigrateTable struct {
}

func (*_sqliteMigrateTable) CreateTableQuery() []string {
	return []string{
		"CREATE TABLE `__migrations__` (" +
			"`id` int unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY," +
			"`name` text NOT NULL," +
			"`timestamp` int unsigned NOT NULL" +
			") ENGINE='InnoDB';",
	}
}

func (*_sqliteMigrateTable) CheckTableQuery() string {
	return "SHOW TABLES LIKE '__migrations__';"
}

func (*_sqliteMigrateTable) CompletedQuery() string {
	return "SELECT `name` FROM `__migrations__`;"
}

func (*_sqliteMigrateTable) SaveQuery() string {
	return "INSERT INTO `__migrations__` (`name`, `timestamp`) VALUES (?, ?);"
}
