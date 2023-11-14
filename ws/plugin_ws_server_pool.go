/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ws

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/ws/server"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

func WithWebsocketServerPool(options ...func(websocket.Upgrader)) plugins.Plugin {
	return plugins.Plugin{
		Inject: func(l xlog.Logger) (*wssPool, WebsocketServerPool) {
			wssp := &wssPool{
				options: options,
				pool:    make(map[string]*server.Server, 10),
				log:     l,
			}
			return wssp, wssp
		},
	}
}

type (
	wssPool struct {
		options []func(websocket.Upgrader)
		pool    map[string]*server.Server
		log     xlog.Logger
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
	p := server.New(v.log, v.ctx, v.options...)
	v.pool[name] = p
	return p
}

func (v *wssPool) Up(ctx xc.Context) error {
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
