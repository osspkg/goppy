/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ws

import (
	"context"
	"net/http"

	"go.osspkg.com/goppy/v3/ws/event"
)

type (
	EventHandler func(event event.Event, meta Meta) error

	Guard func(cid string, head http.Header) error

	Meta interface {
		ConnectID() string
		Head(key string) string
		AddOnCloseFunc(cb func(cid string))
		AddOnOpenFunc(cb func(cid string))
		Context() context.Context
	}
)
