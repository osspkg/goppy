/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package rpc

import (
	"context"
	"fmt"
	"plugin"

	"go.osspkg.com/do"
	"go.osspkg.com/logx"
)

type goPlugin struct {
	conf      Config
	funcStart func(ctx context.Context, opts map[string]string) error
	funcStop  func() error
	funcCall  func(ctx context.Context, method string, params, result any) error
}

func newGoPlugin(c Config) (rpcPlugin, error) {
	obj := &goPlugin{conf: c}

	p, err := plugin.Open(c.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin %s: %w", c.Path, err)
	}

	obj.funcStart, err = lookupFunc[func(context.Context, map[string]string) error](p, "Start")
	if err != nil {
		return nil, fmt.Errorf("failed to lookup function Start in %s: %w", c.Name, err)
	}

	obj.funcStop, err = lookupFunc[func() error](p, "Stop")
	if err != nil {
		return nil, fmt.Errorf("failed to lookup function Stop in %s: %w", c.Name, err)
	}

	obj.funcCall, err = lookupFunc[func(context.Context, string, any, any) error](p, "Call")
	if err != nil {
		return nil, fmt.Errorf("failed to lookup function Call in %s: %w", c.Name, err)
	}

	return obj, nil
}

func (p *goPlugin) Start(ctx context.Context, opts map[string]string) error {
	return p.funcStart(ctx, do.JoinMap(p.conf.Options, opts))
}

func (p *goPlugin) Stop() error {
	return p.funcStop()
}

func (p *goPlugin) Call(ctx context.Context, method string, params, result any) error {
	defer func() {
		if err := recover(); err != nil {
			logx.Error(
				"Go Plugin",
				"err", fmt.Errorf("panic: %v", err),
				"name", p.conf.Name,
			)
		}
	}()
	return p.funcCall(ctx, method, params, result)
}

func lookupFunc[T any](p *plugin.Plugin, funcName string) (callback T, err error) {
	v, err := p.Lookup(funcName)
	if err != nil {
		return callback, fmt.Errorf("failed to lookup %s: %w", funcName, err)
	}

	var ok bool
	if callback, ok = v.(T); !ok {
		return callback, fmt.Errorf("plugin does not implement %s", funcName)
	}

	return callback, nil
}
