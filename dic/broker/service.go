/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package broker

import (
	"context"

	"go.osspkg.com/errors"
	"go.osspkg.com/logx"
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v3/plugins"
)

var ErrServiceUnknown = errors.New("unknown service")

type (
	IService interface {
		Up() error
		Down() error
	}
	IServiceXContext interface {
		Up(ctx xc.Context) error
		Down() error
	}
	IServiceContext interface {
		Up(ctx context.Context) error
		Down() error
	}
)

func isService(arg any) bool {
	if arg == nil {
		return false
	}
	if _, ok := arg.(IServiceContext); ok {
		return true
	}
	if _, ok := arg.(IServiceXContext); ok {
		return true
	}
	if _, ok := arg.(IService); ok {
		return true
	}
	return false
}

func callUp(v interface{}, c xc.Context) error {
	if vv, ok := v.(IServiceContext); ok {
		return vv.Up(c.Context())
	}
	if vv, ok := v.(IServiceXContext); ok {
		return vv.Up(c)
	}
	if vv, ok := v.(IService); ok {
		return vv.Up()
	}
	return errors.Wrapf(ErrServiceUnknown, "service [%T]", v)
}

func callDown(v interface{}) error {
	if vv, ok := v.(IServiceContext); ok {
		return vv.Down()
	}
	if vv, ok := v.(IServiceXContext); ok {
		return vv.Down()
	}
	if vv, ok := v.(IService); ok {
		return vv.Down()
	}
	return errors.Wrapf(ErrServiceUnknown, "service [%T]", v)
}

type _serviceBroker struct {
	objects []interface{}
	index   int
}

func WithServiceBroker() plugins.Broker {
	return &_serviceBroker{
		objects: make([]interface{}, 0, 10),
	}
}

func (s *_serviceBroker) Name() string {
	return "service"
}

func (s *_serviceBroker) Priority() int {
	return -100
}

func (s *_serviceBroker) Apply(arg any) {
	if !isService(arg) {
		return
	}
	s.objects = append(s.objects, arg)
}

func (s *_serviceBroker) OnStart(ctx xc.Context) error {
	logx.Info("Service Broker", "do", "start", "count", len(s.objects))

	if len(s.objects) == 0 {
		return nil
	}

	for i := 0; i < len(s.objects); i++ {
		if err := callUp(s.objects[i], ctx); err != nil {
			return err
		}
		s.index = i
	}

	return nil
}

func (s *_serviceBroker) OnStop() error {
	logx.Info("Service Broker", "do", "stop", "count", len(s.objects))

	if len(s.objects) == 0 {
		return nil
	}

	var errResult error
	for ; s.index >= 0; s.index-- {
		if err := callDown(s.objects[s.index]); err != nil {
			errResult = errors.Wrap(
				errResult,
				errors.Wrapf(err, "down [%T] service error", s.objects[s.index]),
			)
		}
	}

	return errResult
}
