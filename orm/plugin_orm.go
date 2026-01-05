/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"fmt"
	"reflect"

	"go.osspkg.com/do"
	"go.osspkg.com/errors"
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v3/orm/dialect"
	"go.osspkg.com/goppy/v3/plugins"
)

func WithORM(dialectNames ...dialect.Name) plugins.Kind {
	dialectNames = do.Unique(dialectNames)

	var (
		err  error
		kind plugins.Kind
	)

	cfgs := make([]dialect.ConfigInterface, 0, len(dialectNames))
	links := make(map[dialect.Name]dialect.ConfigInterface)

	if len(dialectNames) == 0 {
		err = errors.New("dialect names cannot be empty")
	}

	for _, name := range dialectNames {
		c, ok := dialect.GetConnector(name)
		if !ok {
			err = errors.Wrap(err, fmt.Errorf("dialect '%s' not found", name))
			continue
		}

		if cfg := c.EmptyConfig(); cfg != nil {
			rv := reflect.ValueOf(cfg)
			if rv.IsNil() {
				continue
			}
			if rv.Kind() != reflect.Ptr {
				err = errors.Wrap(err, fmt.Errorf("dialect '%s' config is not a pointer", name))
			}
			cfgs = append(cfgs, cfg)
			links[name] = cfg
		}
	}

	kind.Config = cfgs

	kind.Inject = func(ctx xc.Context) (ORM, error) {
		if err != nil {
			return nil, fmt.Errorf("orm: %w", err)
		}

		obj := New(ctx.Context())
		go func() {
			defer obj.Close()

			<-ctx.Done()
		}()

		for name, cfg := range links {
			if err = obj.ApplyConfig(name, cfg); err != nil {
				return nil, fmt.Errorf("orm: apply config '%s': %w", name, err)
			}

		}

		return obj, nil
	}

	return kind
}
