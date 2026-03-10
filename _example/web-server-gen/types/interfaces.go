/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package types

import "context"

// Api
// @wsg description="Методы апи"
// @wsg web-pool=main,admin
// @wsg route-prefix=/api/v1
type Api interface {
	// Root
	// @wsg in.userID=cookie:x-user-id
	Root(
		ctx context.Context,
		userID int64,
		userName string,
	) (status bool, err error)

	// Auth
	// @wsg in.userID=cookie:x-user-id
	// @wsg module=authz
	Auth(
		ctx context.Context,
		userID int64,
		userName string,
	) (status bool, err error)
}
