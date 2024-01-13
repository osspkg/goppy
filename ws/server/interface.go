/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package server

import "go.osspkg.com/goppy/ws/event"

type (
	EventHandler func(w Response, r Request, m Meta) error

	Request interface {
		EventID() event.Id
		Decode(in interface{}) error
	}

	Response interface {
		Encode(in interface{})
		EncodeEvent(id event.Id, in interface{})
	}

	Meta interface {
		ConnectID() string
		Head(key string) string
		OnClose(cb func(cid string))
		OnOpen(cb func(cid string))
	}
)
