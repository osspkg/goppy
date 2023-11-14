/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package client

import "go.osspkg.com/goppy/ws/event"

type (
	Handler func(w Request, r Response, m Meta)

	Response interface {
		EventID() event.Id
		Decode(in interface{}) error
	}

	Request interface {
		Encode(in interface{})
		EncodeEvent(id event.Id, in interface{})
	}

	Meta interface {
		ConnectID() string
	}
)
