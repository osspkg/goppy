/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package mysql

type migrate struct {
}

func (migrate) CreateTableQuery() []string {
	return []string{
		"CREATE TABLE `__migrations__` (" +
			"`id` int unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY," +
			"`name` text NOT NULL," +
			"`timestamp` int unsigned NOT NULL" +
			") ENGINE='InnoDB';",
	}
}

func (migrate) CheckTableQuery() string {
	return "SHOW TABLES LIKE '__migrations__';"
}

func (migrate) CompletedQuery() string {
	return "SELECT `name` FROM `__migrations__`;"
}

func (migrate) SaveQuery() string {
	return "INSERT INTO `__migrations__` (`name`, `timestamp`) VALUES (?, ?);"
}
