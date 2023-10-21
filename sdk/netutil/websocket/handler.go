/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package websocket

type (
	ClientHandler func(w CRequest, r CResponse, m CMeta)

	CResponse interface {
		EventID() EventID
		Decode(in interface{}) error
	}

	CRequest interface {
		Encode(in interface{})
		EncodeEvent(id EventID, in interface{})
	}

	CMeta interface {
		ConnectID() string
	}
)

type (
	EventHandler func(w Response, r Request, m Meta) error

	Request interface {
		EventID() EventID
		Decode(in interface{}) error
	}

	Response interface {
		Encode(in interface{})
		EncodeEvent(id EventID, in interface{})
	}

	Meta interface {
		ConnectID() string
		Head(key string) string
		OnClose(cb func(cid string))
		OnOpen(cb func(cid string))
	}
)
