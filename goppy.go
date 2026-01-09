/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package goppy

import (
	"context"

	"go.osspkg.com/config"
	cenv "go.osspkg.com/config/env"
	"go.osspkg.com/events"
	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"
	"go.osspkg.com/xc"
	"go.uber.org/automaxprocs/maxprocs"

	"go.osspkg.com/goppy/v3/internal/appsteps"

	"go.osspkg.com/goppy/v3/console"
	"go.osspkg.com/goppy/v3/dic"
	"go.osspkg.com/goppy/v3/dic/broker"
	"go.osspkg.com/goppy/v3/env"
	"go.osspkg.com/goppy/v3/internal/appconfig"
	"go.osspkg.com/goppy/v3/internal/applog"
	"go.osspkg.com/goppy/v3/internal/appreflect"
	"go.osspkg.com/goppy/v3/plugins"
)

func init() {
	_, err := maxprocs.Set()
	console.FatalIfErr(err, "set auto max process")
}

type (
	_app struct {
		info      *env.AppInfo
		container *dic.Container
		console   *console.Console
		configs   []any
		resolvers []config.Resolver
	}

	Goppy interface {
		Plugins(args ...any)
		Command(cb func(console.CommandSetter))
		Run()
	}
)

// New constructor for init Goppy
func New(name, version, description string) Goppy {
	return &_app{
		info: func() *env.AppInfo {
			info := env.NewAppInfo()
			info.AppName = env.AppName(name)
			info.AppVersion = env.AppVersion(version)
			info.AppDescription = env.AppDescription(description)
			return &info
		}(),
		container: dic.New(),
		console:   console.New(name, description),
		configs:   make([]any, 0, 10),
		resolvers: make([]config.Resolver, 0, 10),
	}
}

// Plugins setting the list of plugins to initialize
func (v *_app) Plugins(dependency ...any) {
	args := plugins.Inject(dependency...)

	for _, arg := range args {

		for _, item := range appreflect.AnySlice(arg.Config) {
			appreflect.Validate(item, plugins.AllowedKindConfig(), func(in any) error {
				v.configs = append(v.configs, in)
				return nil
			})
		}

		for _, item := range appreflect.AnySlice(arg.Inject) {
			appreflect.Validate(item, plugins.AllowedKindInject(), func(in any) error {
				switch val := in.(type) {
				case plugins.Broker:
					return v.container.BrokerRegister(val)
				case config.Resolver:
					v.resolvers = append(v.resolvers, val)
				default:
					return v.container.Register(in)
				}
				return nil
			})
		}
	}
}

func (v *_app) Command(cb func(console.CommandSetter)) {
	v.console.AddCommand(console.NewCommand(cb))
}

const (
	configInited = "configInited"
	logInited    = "logInited"
	logDone      = "logDone"
	appExit      = "*appExit"
)

// Run launching Goppy with initialization of all dependencies
func (v *_app) Run() {
	ctx := xc.New()
	go events.OnStopSignal(ctx.Close)
	console.FatalIfErr(v.container.Register(
		func() xc.Context { return ctx },
	), "register base dependency")

	steps := appsteps.New(ctx.Context(),
		configInited, logInited, logDone, appExit,
	)

	wg := syncing.NewGroup(ctx.Context())
	wg.OnPanic(func(e error) { logx.Error("Run background", "err", e) })

	{
		conf := &applog.GroupConfig{}
		v.configs = append(v.configs, conf)

		wg.Background("log writer", func(_ context.Context) {
			steps.Wait(configInited)
			lw := applog.New(string(v.info.AppName), conf.Log)
			steps.Done(logInited).Wait(appExit)
			console.WarnIfErr(lw.Close(), "close log file")
			steps.Done(logDone)
		})
	}

	{
		if len(v.resolvers) == 0 {
			v.resolvers = append(v.resolvers, cenv.New())
		}

		console.FatalIfErr(v.container.BrokerRegister(
			broker.WithTickerBroker(),
			broker.WithServiceBroker(),
		), "register default broker")

		console.FatalIfErr(v.container.Register(
			v.info.AppName,
			v.info.AppVersion,
			v.info.AppDescription,
			*v.info,
		), "register app info")
	}

	{
		v.console.AddGlobal(console.NewCommand(func(setter console.CommandSetter) {
			setter.Flag(func(flagsSetter console.FlagsSetter) {
				flagsSetter.StringVar("config", "", "Set config file path")
				flagsSetter.StringVar("config-env", "", "Set config data from env")
				flagsSetter.StringVar("config-ext", ".yaml", "Set config data format")
				flagsSetter.Bool("config-recovery", "Recovery config if empty")
			})
			setter.ExecFunc(func(confFile, confEnv, confExt string, confRecovery bool) {
				conf := appconfig.Config{
					Filepath: confFile,
					Data:     env.Get(confEnv, ""),
					Ext:      confExt,
				}
				if len(confFile) > 0 && confRecovery {
					console.FatalIfErr(appconfig.Recovery(confFile, v.configs), "config recovery")
				}
				console.FatalIfErr(appconfig.DecodeAndValidate(conf, v.resolvers, v.configs), "config validate")
				console.FatalIfErr(v.container.Register(v.configs...), "register config")

				steps.Done(configInited).Wait(logInited)
			})
		}))

		v.console.RootCommand(console.NewCommand(func(setter console.CommandSetter) {
			setter.Setup(string(v.info.AppName), string(v.info.AppDescription))
			setter.Flag(func(flagsSetter console.FlagsSetter) {
				flagsSetter.StringVar("pid", "", "Set PID file path")
			})
			setter.ExecFunc(func(pid string) {
				if len(pid) > 0 {
					console.FatalIfErr(appconfig.CreatePID(pid), "create pid file")
				}

				console.FatalIfErr(v.container.Start(ctx), "start dependency")
				<-ctx.Done()
				console.WarnIfErr(v.container.Stop(), "stop dependency")
			})
		}))
	}

	v.console.Exec()
	steps.Done(appExit).Wait(logDone)
	wg.Wait()
}
