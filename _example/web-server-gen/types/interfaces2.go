/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package types

import "context"

type Post interface {
	ByID(
		ctx context.Context,
		ID int64,
	) (text bool, err error)

	List(
		ctx context.Context,
		userID int64,
	) (text []Text, err error)
}

type Text struct {
	ID   int64
	Text string
}
