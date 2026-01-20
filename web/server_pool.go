/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"fmt"

	"go.osspkg.com/errors"
	"go.osspkg.com/xc"
)

type (
	// ServerPool router pool handler
	ServerPool interface {
		All(call func(tag string, router Router))
		ByTag(tag string) (Router, bool)
		Main() (Router, bool)
		Admin() (Router, bool)
	}

	serverPool struct {
		pool map[string]*route
	}
)

func newServerPool(configs []Config) (*serverPool, error) {
	v := &serverPool{
		pool: make(map[string]*route),
	}

	for _, config := range configs {
		if _, ok := v.pool[config.Tag]; ok {
			return nil, fmt.Errorf("http server pool: duplicate tag: %s", config.Tag)
		}
		v.pool[config.Tag] = newRouter(config.Tag, config)
	}

	return v, nil
}

// All method to get all route handlers
func (v *serverPool) All(call func(name string, router Router)) {
	for n, r := range v.pool {
		call(n, r)
	}
}

// Main method to get Main route handler
func (v *serverPool) Main() (Router, bool) {
	return v.ByTag("main")
}

// Admin method to get Admin route handler
func (v *serverPool) Admin() (Router, bool) {
	return v.ByTag("admin")
}

// ByTag method to get route handler by tag
func (v *serverPool) ByTag(name string) (Router, bool) {
	if r, ok := v.pool[name]; ok {
		return r, true
	}
	return nil, false
}

func (v *serverPool) Up(c xc.Context) error {
	var err error
	for n, r := range v.pool {
		err = errors.Wrap(err, errors.Wrapf(r.Up(c), "tag='%s'", n))
	}

	return errors.Wrapf(err, "http server pool: failed start")
}

func (v *serverPool) Down() error {
	var err error
	for n, r := range v.pool {
		err = errors.Wrap(err, errors.Wrapf(r.Down(), "tag='%s'", n))
	}

	return errors.Wrapf(err, "http server pool: failed stop")
}
