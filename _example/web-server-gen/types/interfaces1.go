/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package types

import "context"

type Api interface {
	// Root
	// @tb in.userID=cookie:x-user-id
	Root(
		ctx context.Context,
		userID int64,
		userName string,
	) (status bool, err error)

	// Auth
	// @tb in.userID=header:x-user-id
	// @tb out.status=header:x-user-id,cookie:uid
	// @tb out.status=cookie:uid
	Auth(
		ctx context.Context,
		userID int64,
		userName string,
	) (status bool, err error)
}

type User interface {
	Name(
		ctx context.Context,
		userID int64,
	) (name string, err error)
}
