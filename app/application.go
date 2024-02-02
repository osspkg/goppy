/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app

import (
	"go.osspkg.com/goppy/console"
	"go.osspkg.com/goppy/env"
	"go.osspkg.com/goppy/iofile"
	"go.osspkg.com/goppy/syscall"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

type (
	App interface {
		Logger(log xlog.Logger) App
		Modules(modules ...interface{}) App
		ConfigFile(filename string, configs ...interface{}) App
		PidFile(filename string) App
		Run()
		Invoke(call interface{})
		Call(call interface{})
		ExitFunc(call func(code int)) App
	}

	_app struct {
		configFilePath string
		pidFilePath    string
		configs        Modules
		modules        Modules
		sources        iofile.FileCodec
		packages       Container
		logHandler     *_log
		log            xlog.Logger
		appContext     xc.Context
		exitFunc       func(code int)
	}
)

// New create application
func New() App {
	ctx := xc.New()
	return &_app{
		modules:    Modules{},
		configs:    Modules{},
		packages:   NewContainer(ctx),
		appContext: ctx,
		exitFunc:   func(_ int) {},
	}
}

// Logger setup logger
func (a *_app) Logger(log xlog.Logger) App {
	a.log = log
	return a
}

// Modules append object to modules list
func (a *_app) Modules(modules ...interface{}) App {
	for _, mod := range modules {
		switch v := mod.(type) {
		case Modules:
			a.modules = a.modules.Add(v...)
		default:
			a.modules = a.modules.Add(v)
		}
	}

	return a
}

// ConfigFile set config file path and configs models
func (a *_app) ConfigFile(filename string, configs ...interface{}) App {
	a.configFilePath = filename
	for _, config := range configs {
		a.configs = a.configs.Add(config)
	}

	return a
}

func (a *_app) PidFile(filename string) App {
	a.pidFilePath = filename
	return a
}

func (a *_app) ExitFunc(v func(code int)) App {
	a.exitFunc = v
	return a
}

// Run application with all dependencies
func (a *_app) Run() {
	a.prepareConfig(false)

	result := a.steps(
		[]step{
			{
				Message: "Registering dependencies",
				Call:    func() error { return a.packages.Register(a.modules...) },
			},
			{
				Message: "Running dependencies",
				Call:    func() error { return a.packages.Start() },
			},
		},
		func(er bool) {
			if er {
				a.appContext.Close()
				return
			}
			go syscall.OnStop(a.appContext.Close)
			<-a.appContext.Done()
		},
		[]step{
			{
				Message: "Stop dependencies",
				Call:    func() error { return a.packages.Stop() },
			},
		},
	)
	console.FatalIfErr(a.logHandler.Close(), "close log file")
	if result {
		a.exitFunc(1)
	}
	a.exitFunc(0)
}

// Invoke run application with all dependencies and call function after starting
func (a *_app) Invoke(call interface{}) {
	a.prepareConfig(true)

	result := a.steps(
		[]step{
			{
				Call: func() error { return a.packages.Register(a.modules...) },
			},
			{
				Call: func() error { return a.packages.Start() },
			},
			{
				Call: func() error { return a.packages.Invoke(call) },
			},
		},
		func(_ bool) {},
		[]step{
			{
				Call: func() error { return a.packages.Stop() },
			},
		},
	)
	console.FatalIfErr(a.logHandler.Close(), "close log file")
	if result {
		a.exitFunc(1)
	}
	a.exitFunc(0)
}

// Call function with dependency and without starting all app
func (a *_app) Call(call interface{}) {
	a.prepareConfig(true)

	result := a.steps(
		[]step{
			{
				Call: func() error { return a.packages.Register(a.modules...) },
			},
			{
				Call: func() error { return a.packages.Register(call) },
			},
			{
				Call: func() error { return a.packages.BreakPoint(call) },
			},
			{
				Call: func() error { return a.packages.Start() },
			},
		},
		func(_ bool) {},
		[]step{
			{
				Call: func() error { return a.packages.Stop() },
			},
		},
	)
	console.FatalIfErr(a.logHandler.Close(), "close log file")
	if result {
		a.exitFunc(1)
	}
	a.exitFunc(0)
}

func (a *_app) prepareConfig(interactive bool) {
	var err error
	if len(a.configFilePath) == 0 {
		a.logHandler = newLog(LogConfig{
			Level:    4,
			FilePath: "/dev/stdout",
			Format:   "string",
		})
		if a.log == nil {
			a.log = xlog.Default()
		}
		a.logHandler.Handler(a.log)
	}
	if len(a.configFilePath) > 0 {
		// read config file
		a.sources = iofile.FileCodec(a.configFilePath)

		// init logger
		config := &Config{}
		if err = a.sources.Decode(config); err != nil {
			console.FatalIfErr(err, "decode config file: %s", a.configFilePath)
		}
		if interactive {
			config.Log.Level = 4
			config.Log.FilePath = "/dev/stdout"
		}
		a.logHandler = newLog(config.Log)
		if a.log == nil {
			a.log = xlog.Default()
		}
		a.logHandler.Handler(a.log)
		a.modules = a.modules.Add(
			env.ENV(config.Env),
		)
		// decode all configs
		var configs []interface{}
		configs, err = typingReflectPtr(a.configs, func(i interface{}) error {
			return a.sources.Decode(i)
		})
		if err != nil {
			a.log.WithFields(xlog.Fields{
				"err": err.Error(),
			}).Fatalf("Decode config file")
		}
		a.modules = a.modules.Add(configs...)

		if !interactive && len(a.pidFilePath) > 0 {
			if err = syscall.Pid(a.pidFilePath); err != nil {
				a.log.WithFields(xlog.Fields{
					"err":  err.Error(),
					"file": a.pidFilePath,
				}).Fatalf("Create pid file")
			}
		}
	}
	a.modules = a.modules.Add(
		func() xlog.Logger { return a.log },
		func() xc.Context { return a.appContext },
	)
}

type step struct {
	Call    func() error
	Message string
}

func (a *_app) steps(up []step, wait func(bool), down []step) bool {
	var erc int

	for _, s := range up {
		if len(s.Message) > 0 {
			a.log.Infof(s.Message)
		}
		if err := s.Call(); err != nil {
			a.log.WithFields(xlog.Fields{
				"err": err.Error(),
			}).Errorf(s.Message)
			erc++
			break
		}
	}

	wait(erc > 0)

	for _, s := range down {
		if len(s.Message) > 0 {
			a.log.Infof(s.Message)
		}
		if err := s.Call(); err != nil {
			a.log.WithFields(xlog.Fields{
				"err": err.Error(),
			}).Errorf(s.Message)
			erc++
		}
	}

	return erc > 0
}
