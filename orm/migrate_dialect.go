/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"fmt"

	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"
)

var (
	migrateHandlers = make([]func(dialect string), 0)
	migrateQuery    = make(map[string]Migrator, 10)
	migrateMux      = syncing.NewLock()
)

func dialectRegister(dialect string, migrate Migrator) {
	migrateMux.Lock(func() {
		migrateQuery[dialect] = migrate
		logx.Debug("Register DB dialect", "dialect", dialect)
		for _, handler := range migrateHandlers {
			go handler(dialect)
		}
	})
}

func dialectExtract(dialect string) (migrate Migrator, err error) {
	migrateMux.Lock(func() {
		m, ok := migrateQuery[dialect]
		if ok {
			migrate = m
			return
		}
		err = fmt.Errorf("migrate dialect [%s] not inited", dialect)
	})
	return
}

func dialectOnRegistered(call func(dialect string)) {
	exists := make([]string, 0)
	migrateMux.Lock(func() {
		for dialect := range migrateQuery {
			exists = append(exists, dialect)
		}
		migrateHandlers = append(migrateHandlers, call)
	})
	for _, dialect := range exists {
		go call(dialect)
	}
	return
}
