/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package broker

import (
	"context"
	"fmt"
	"time"

	"go.osspkg.com/errors"
	"go.osspkg.com/logx"
	"go.osspkg.com/routine/tick"
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v3/plugins"
)

type ITickConfig interface {
	ApplyTickConfig(call func(tick.Config))
}

type _tickerBroker struct {
	tik    *tick.Ticker
	ctx    context.Context
	cancel context.CancelFunc
}

func WithTickerBroker() plugins.Broker {
	ctx, cancel := context.WithCancel(context.Background())
	return &_tickerBroker{
		tik: &tick.Ticker{
			Calls: make([]tick.Config, 0, 4),
			OnError: func(name string, err error) {
				logx.Error("Time Ticker Broker", "name", name, "err", err)
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (i *_tickerBroker) Name() string {
	return "interval start-up"
}

func (i *_tickerBroker) Priority() int {
	return -99
}

func (i *_tickerBroker) Apply(arg any) {
	obj, ok := arg.(ITickConfig)
	if !ok {
		return
	}
	obj.ApplyTickConfig(func(c tick.Config) {
		i.tik.Calls = append(i.tik.Calls, tick.Config{
			Name:     c.Name,
			OnStart:  c.OnStart,
			Interval: c.Interval,
			Func: func(ctx context.Context, t time.Time) (err error) {
				defer func() {
					if r := recover(); r != nil {
						err = errors.Wrap(err, fmt.Errorf("[PANIC] %w", fmt.Errorf("%+v", r)))
					}
				}()
				err = c.Func(ctx, t)
				return
			},
		})
	})
}

func (i *_tickerBroker) OnStart(ctx xc.Context) error {
	logx.Info("Time Ticker Broker", "do", "start", "count", len(i.tik.Calls))

	go func() {
		i.tik.Run(ctx.Context())
		i.cancel()
	}()
	return nil
}

func (i *_tickerBroker) OnStop() error {
	logx.Info("Time Ticker Broker", "do", "stop", "count", len(i.tik.Calls))
	<-i.ctx.Done()
	return nil
}
