/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dialect

import "go.osspkg.com/syncing"

var clients = syncing.NewMap[Name, Connector](3)
var migrates = syncing.NewMap[Name, Migrator](3)

func GetConnector(name Name) (Connector, bool) {
	return clients.Get(name)
}

func GetMigrator(name Name) (Migrator, bool) {
	return migrates.Get(name)
}

func Register(name Name, connector Connector, migrator Migrator) {
	clients.Set(name, connector)
	migrates.Set(name, migrator)
}
