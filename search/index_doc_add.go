/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search

import (
	"fmt"
	"time"
)

func (v *service) AddDocuments(name, id string, data ...map[string]string) error {
	if len(data) == 0 {
		return nil
	}

	index, ok := v.list.Get(name)
	if !ok {
		return fmt.Errorf("no such index '%s'", name)
	}

	curr := time.Now().UTC().Format(time.RFC3339)
	batch := index.NewBatch()

	for _, datum := range data {
		datum[fieldCreatedAt] = curr
		if err := batch.Index(id, datum); err != nil {
			return fmt.Errorf("add document '%+v': %w", datum, err)
		}
	}

	return index.Batch(batch)
}
