/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search

import (
	"fmt"
)

func (v *service) DeleteDocuments(name string, ids ...string) error {
	if len(ids) == 0 {
		return nil
	}

	index, ok := v.list.Get(name)
	if !ok {
		return fmt.Errorf("no such index '%s'", name)
	}

	batch := index.NewBatch()

	for _, id := range ids {
		batch.Delete(id)
	}

	return index.Batch(batch)
}
