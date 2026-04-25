/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package rpc

import (
	"context"
	"fmt"

	"go.osspkg.com/logx"
)

type RPC struct {
	conf    []Config
	plugins map[string]rpcPlugin
}

func New(conf ...Config) *RPC {
	return &RPC{
		conf:    conf,
		plugins: make(map[string]rpcPlugin, len(conf)),
	}
}

func (v *RPC) Up(ctx context.Context) error {
	for _, c := range v.conf {
		if _, ok := v.plugins[c.Name]; ok {
			return fmt.Errorf("plugin %s already exists", c.Name)
		}

		var (
			p   rpcPlugin
			err error
		)
		switch c.Type {
		case TypeGoPlugin:
			p, err = newGoPlugin(c)
		case TypeUnixSocket:
			p, err = newUnixPlugin(c)
		default:
			return fmt.Errorf("unknown plugin type: %s (name=%s, path=%s)", c.Type, c.Name, c.Path)
		}

		if err != nil {
			return fmt.Errorf("failed to init plugin: %w (name=%s, path=%s)", err, c.Name, c.Path)
		}

		if err = p.Start(ctx, nil); err != nil {
			logx.Error("failed to start rpc plugin", "err", err, "name", c.Name, "path", c.Path)
			continue
		}

		v.plugins[c.Name] = p
	}
	return nil
}

func (v *RPC) Down() error {
	for name, p := range v.plugins {
		if err := p.Stop(); err != nil {
			logx.Error("failed to stop rpc plugin", "err", err, "name", name)
		}
	}

	return nil
}

func (v *RPC) Call(ctx context.Context, name string, method string, params, result any) error {
	if p, ok := v.plugins[name]; ok {
		return p.Call(ctx, method, params, result)
	}
	return fmt.Errorf("plugin %s not found", name)
}
