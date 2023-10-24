/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"context"
	"sync"

	ws "github.com/gorilla/websocket"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/sdk/app"
	"go.osspkg.com/goppy/sdk/log"
	"go.osspkg.com/goppy/sdk/netutil/websocket"
)

func WithWebsocketServerPool(options ...func(ws.Upgrader)) plugins.Plugin {
	return plugins.Plugin{
		Inject: func(l log.Logger) (*wssPool, WebsocketServerPool) {
			wssp := &wssPool{
				options: options,
				pool:    make(map[string]*websocket.Server, 10),
				log:     l,
			}
			return wssp, wssp
		},
	}
}

type (
	wssPool struct {
		options []func(ws.Upgrader)
		pool    map[string]*websocket.Server
		log     log.Logger
		ctx     context.Context
		mux     sync.Mutex
	}

	WebsocketServerPool interface {
		Create(name string) WebsocketServer
	}
)

func (v *wssPool) Create(name string) WebsocketServer {
	v.mux.Lock()
	defer v.mux.Unlock()

	if p, ok := v.pool[name]; ok {
		return p
	}
	p := websocket.NewServer(v.log, v.ctx, v.options...)
	v.pool[name] = p
	return p
}

func (v *wssPool) Up(ctx app.Context) error {
	v.ctx = ctx.Context()
	return nil
}

func (v *wssPool) Down() error {
	v.mux.Lock()
	defer v.mux.Unlock()

	for _, item := range v.pool {
		item.CloseAll()
	}

	return nil
}
