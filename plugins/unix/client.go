/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unix

import (
	"sync"

	"github.com/osspkg/goppy/plugins"
	"github.com/osspkg/goppy/sdk/netutil/unixsocket"
)

func WithClient() plugins.Plugin {
	return plugins.Plugin{
		Inject: func() (*clientProvider, Client) {
			s := newClientProvider()
			return s, s
		},
	}
}

type (
	clientProvider struct {
		list map[string]ClientConnect
		mux  sync.RWMutex
	}

	Client interface {
		Create(path string) (ClientConnect, error)
	}

	ClientConnect interface {
		Exec(name string, b []byte) ([]byte, error)
		ExecString(name string, b string) ([]byte, error)
	}
)

func newClientProvider() *clientProvider {
	return &clientProvider{
		list: make(map[string]ClientConnect),
	}
}

func (v *clientProvider) Create(path string) (ClientConnect, error) {
	v.mux.Lock()
	defer v.mux.Unlock()
	if c, ok := v.list[path]; ok {
		return c, nil
	}
	c := unixsocket.NewClient(path)
	v.list[path] = c
	return c, nil
}
