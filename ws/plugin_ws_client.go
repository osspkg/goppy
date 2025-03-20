/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ws

import (
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/ws/event"
)

type Client interface {
	Open(url string, opts ...func(Option)) (string, error)
	SetEventHandler(h EventHandler, eids ...event.Id)
	DelEventHandler(eids ...event.Id)
	CountConn() (cc int)
	OnClose(cb func(cid string))
	OnOpen(cb func(cid string))
	CloseOne(cid string)
	CloseAll()
	BroadcastEvent(eid event.Id, m any) (err error)
	SendEvent(eid event.Id, m any, cids ...string) (err error)
}

func WithClient() plugins.Plugin {
	return plugins.Plugin{
		Inject: func(ctx xc.Context) Client {
			return NewClient(ctx.Context())
		},
	}
}
