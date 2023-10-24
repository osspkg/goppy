/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package websocket

import (
	"context"
	"sync/atomic"
	"time"

	"go.osspkg.com/goppy/sdk/iosync"
)

type (
	Observable interface {
		Subscribe(eid EventID, in interface{}) Subscription
	}

	ObservableClient interface {
		OnClose(cb func(cid string))
		Encode(eid EventID, in interface{})
		DelHandler(eids ...EventID)
		SetHandler(call ClientHandler, eids ...EventID)
	}

	_obs struct {
		cli ObservableClient
	}
)

func NewObservable(cli ObservableClient) Observable {
	return &_obs{
		cli: cli,
	}
}

func (v *_obs) Subscribe(eid EventID, in interface{}) Subscription {
	ctx, cncl := context.WithCancel(context.TODO())
	sub := &_sub{
		eid:   eid,
		count: 0,
		ctx:   ctx,
		cncl:  cncl,
		cli:   v.cli,
		call: func() {
			if in == nil {
				return
			}
			v.cli.Encode(eid, in)
		},
		sync: iosync.NewSwitch(),
	}
	v.cli.OnClose(func(_ string) {
		sub.Unsubscribe()
	})
	return sub
}

type (
	cliApi interface {
		DelHandler(eids ...EventID)
		SetHandler(call ClientHandler, eids ...EventID)
	}
	_sub struct {
		eid      EventID
		count    uint64
		maxCount uint64
		ctx      context.Context
		cncl     context.CancelFunc
		cli      cliApi
		call     func()
		sync     iosync.Switch
	}

	Subscription interface {
		Listen(call func(ListenArg), pipe ...PipeFunc)
		Unsubscribe()
	}

	ListenArg interface {
		Decode(in interface{}) error
	}
)

func (v *_sub) Listen(call func(ListenArg), pipe ...PipeFunc) {
	if !v.sync.On() {
		return
	}
	for _, fn := range pipe {
		v.ctx = fn(v.ctx)
	}
	if tc, ok := v.ctx.Value(pipeTakeKey).(uint64); ok {
		v.maxCount = tc
	}
	v.cli.SetHandler(func() func(w CRequest, r CResponse, m CMeta) {
		return func(_ CRequest, r CResponse, _ CMeta) {
			atomic.AddUint64(&v.count, 1)
			call(r)
			if v.maxCount > 0 && atomic.LoadUint64(&v.count) >= v.maxCount {
				v.Unsubscribe()
			}
		}
	}(), v.eid)
	v.call()
	<-v.ctx.Done()
	v.cli.DelHandler(v.eid)
}

func (v *_sub) Unsubscribe() {
	v.cncl()
}

type (
	PipeFunc func(ctx context.Context) context.Context
	pipeKey  string
)

var (
	pipeTakeKey pipeKey = "take"
)

func PipeTimeout(t time.Duration) PipeFunc {
	return func(ctx context.Context) context.Context {
		c, _ := context.WithTimeout(ctx, t) //nolint: govet
		return c
	}
}

func PipeTake(count uint64) PipeFunc {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, pipeTakeKey, count)
	}
}
