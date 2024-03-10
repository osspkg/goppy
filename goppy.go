/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package goppy

import (
	"fmt"
	"os"
	"reflect"

	"go.osspkg.com/goppy/app"
	"go.osspkg.com/goppy/config"
	"go.osspkg.com/goppy/console"
	"go.osspkg.com/goppy/env"
	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/xlog"
	"gopkg.in/yaml.v3"
)

type (
	_app struct {
		application app.App
		commands    map[string]interface{}
		plugins     []interface{}
		configs     []interface{}
		resolvers   []config.Resolver
		args        *console.Args
		info        *env.AppInfo
	}

	Goppy interface {
		AppName(t string)
		AppVersion(t string)
		AppDescription(t string)
		Logger(l xlog.Logger)
		Plugins(args ...plugins.Plugin)
		Command(name string, call interface{})
		ConfigResolvers(rc ...config.Resolver)
		Run()
	}
)

// New constructor for init Goppy
func New() Goppy {
	return &_app{
		application: app.New().ExitFunc(func(code int) {
			os.Exit(code)
		}),
		commands:  make(map[string]interface{}),
		plugins:   make([]interface{}, 0, 100),
		configs:   make([]interface{}, 0, 100),
		resolvers: make([]config.Resolver, 0, 100),
		args:      console.NewArgs().Parse(os.Args[1:]),
		info: func() *env.AppInfo {
			info := env.NewAppInfo()
			return &info
		}(),
	}
}

func (v *_app) AppName(t string) {
	v.info.AppName = env.AppName(t)
}

func (v *_app) AppVersion(t string) {
	v.info.AppVersion = env.AppVersion(t)
}

func (v *_app) AppDescription(t string) {
	v.info.AppDescription = env.AppDescription(t)
}

func (v *_app) Logger(l xlog.Logger) {
	v.application.Logger(l)
}

func (v *_app) ConfigResolvers(rc ...config.Resolver) {
	v.application.ConfigResolvers(rc...)
}

// Plugins setting the list of plugins to initialize
func (v *_app) Plugins(args ...plugins.Plugin) {
	for _, arg := range args {
		reflectResolve(arg.Config, plugins.AllowedKindConfig, func(in interface{}) {
			v.configs = append(v.configs, in)
		})
		reflectResolve(arg.Inject, plugins.AllowedKindInject, func(in interface{}) {
			v.plugins = append(v.plugins, in)
		})
		reflectResolve(arg.Resolve, plugins.AllowedKindResolve, func(in interface{}) {
			v.plugins = append(v.plugins, in)
		})
	}
}

func (v *_app) Command(name string, call interface{}) {
	v.commands[name] = call
}

// Run launching Goppy with initialization of all dependencies
func (v *_app) Run() {
	if len(v.resolvers) == 0 {
		v.ConfigResolvers(config.EnvResolver())
	}

	apps := v.application.Modules(v.plugins...)
	apps.Modules(v.info.AppName, v.info.AppVersion, v.info.AppDescription, *v.info)

	appConfig := v.parseConfigFlag()
	console.FatalIfErr(v.recoveryConfig(appConfig), "config recovery")
	console.FatalIfErr(v.validateConfig(appConfig), "config validate")
	apps.ConfigFile(appConfig)
	apps.ConfigModels(v.configs...)
	apps.ConfigResolvers(v.resolvers...)

	pid, err := v.parsePIDFileFlag()
	console.FatalIfErr(err, "check pid file")
	apps.PidFile(pid)

	if params := v.args.Next(); len(params) > 0 {
		if cmd, ok := v.commands[params[0]]; ok {
			apps.Call(cmd)
			return
		}
		console.Fatalf("<%s> command not found", params[0])
	}
	apps.Run()
}

func reflectResolve(arg interface{}, k plugins.AllowedKind, call func(interface{})) {
	if arg == nil {
		return
	}
	k.MustValidate(arg)
	call(arg)
}

func (v *_app) parseConfigFlag() string {
	conf := v.args.Get("config")
	if conf == nil || len(*conf) == 0 {
		return ""
	}
	return *conf
}

func (v *_app) parsePIDFileFlag() (string, error) {
	pid := v.args.Get("pid")
	if pid == nil || len(*pid) == 0 {
		return "", nil
	}
	file, err := os.Create(*pid)
	if err != nil {
		return "", err
	}
	if err = file.Close(); err != nil {
		return "", err
	}
	return *pid, nil
}

func (v *_app) validateConfig(filename string) error {
	if len(filename) == 0 {
		return nil
	}
	rc := config.New(v.resolvers...)
	if err := rc.OpenFile(filename); err != nil {
		return err
	}
	if err := rc.Build(); err != nil {
		return err
	}
	defType := reflect.TypeOf(new(plugins.Validator)).Elem()
	for _, cfg := range v.configs {
		if reflect.TypeOf(cfg).AssignableTo(defType) {
			if err := rc.Decode(cfg); err != nil {
				return fmt.Errorf("decode config %T error: %w", cfg, err)
			}
			vv, ok := cfg.(plugins.Validator)
			if !ok {
				continue
			}
			if err := vv.Validate(); err != nil {
				return fmt.Errorf("validate config %T error: %w", cfg, err)
			}
		}
	}
	return nil
}

func (v *_app) recoveryConfig(filename string) error {
	if len(filename) == 0 {
		return nil
	}
	_, err := os.Stat(filename)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	b, err := yaml.Marshal(&app.Config{
		Env: "dev",
		Log: app.LogConfig{
			Level:    4,
			FilePath: "/dev/stdout",
			Format:   "string",
		},
	})
	if err != nil {
		return err
	}
	defType := reflect.TypeOf(new(plugins.Defaulter)).Elem()
	for _, cfg := range v.configs {
		if reflect.TypeOf(cfg).AssignableTo(defType) {
			reflect.ValueOf(cfg).MethodByName("Default").Call([]reflect.Value{})
		}
		if bb, err0 := yaml.Marshal(cfg); err0 == nil {
			b = append(b, '\n')
			b = append(b, bb...)
		} else {
			return err0
		}
	}
	if err = os.WriteFile(filename, b, 0755); err != nil {
		return err
	}
	return nil
}
