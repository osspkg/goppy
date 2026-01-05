/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search

import "time"

type Result struct {
	Total     uint64
	Took      time.Duration
	MaxScore  float64
	Documents []Document
}

type Document struct {
	ID        string
	Score     float64
	Fields    map[string]string
	CreatedAt time.Time
}
