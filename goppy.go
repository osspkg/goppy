/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package goppy

import (
	"fmt"
	"os"
	"reflect"

	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/sdk/app"
	"go.osspkg.com/goppy/sdk/console"
	"go.osspkg.com/goppy/sdk/errors"
	"gopkg.in/yaml.v3"
)

type (
	_app struct {
		application app.App
		commands    map[string]interface{}
		plugins     []interface{}
		configs     []interface{}
		args        *console.Args
	}

	Goppy interface {
		Plugins(args ...plugins.Plugin)
		Command(name string, call interface{})
		Run()
	}
)

// New constructor for init Goppy
func New() Goppy {
	return &_app{
		application: app.New().ExitFunc(func(code int) {
			os.Exit(code)
		}),
		commands: make(map[string]interface{}),
		plugins:  make([]interface{}, 0, 100),
		configs:  make([]interface{}, 0, 100),
		args:     console.NewArgs().Parse(os.Args[1:]),
	}
}

// Plugins setting the list of plugins to initialize
func (v *_app) Plugins(args ...plugins.Plugin) {
	for _, arg := range args {
		reflectResolve(arg.Config, reflect.Ptr, func(in interface{}) {
			v.configs = append(v.configs, in)
		}, "Plugin.Config can only be a reference to an object")
		reflectResolve(arg.Inject, reflect.Func, func(in interface{}) {
			v.plugins = append(v.plugins, in)
		}, "Plugin.Inject can only be a function that accepts "+
			"dependencies and returns a reference to the initialized service")
		reflectResolve(arg.Resolve, reflect.Func, func(in interface{}) {
			v.plugins = append(v.plugins, in)
		}, "Plugin.Resolve can only be a function that accepts dependencies")
	}
}

func (v *_app) Command(name string, call interface{}) {
	v.commands[name] = call
}

// Run launching Goppy with initialization of all dependencies
func (v *_app) Run() {
	apps := v.application.Modules(v.plugins...)

	config := v.parseConfigFlag()
	console.FatalIfErr(recoveryConfig(config, v.configs...), "config recovery")
	console.FatalIfErr(validateConfig(config, v.configs...), "config validate")
	apps.ConfigFile(config, v.configs...)

	pid, err := v.parsePIDFileFlag()
	console.FatalIfErr(err, "check pid file")
	apps.PidFile(pid)

	if params := v.args.Next(); len(params) > 0 {
		if cmd, ok := v.commands[params[0]]; ok {
			apps.Invoke(cmd)
			return
		}
		console.Fatalf("<%s> command not found", params[0])
	}
	apps.Run()
}

func reflectResolve(arg interface{}, k reflect.Kind, call func(interface{}), comment string) {
	if arg == nil {
		return
	}
	if reflect.TypeOf(arg).Kind() != k {
		panic(comment)
	}
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

func validateConfig(filename string, configs ...interface{}) error {
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
	defType := reflect.TypeOf(new(plugins.Validator)).Elem()
	for _, cfg := range configs {
		if reflect.TypeOf(cfg).AssignableTo(defType) {
			if err = app.Sources(filename).Decode(cfg); err != nil {
				return fmt.Errorf("decode config %T error: %w", cfg, err)
			}
			vv, ok := cfg.(plugins.Validator)
			if !ok {
				continue
			}
			if err = vv.Validate(); err != nil {
				return fmt.Errorf("validate config %T error: %w", cfg, err)
			}
		}
	}
	return nil
}

func recoveryConfig(filename string, configs ...interface{}) error {
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
		Env:     "dev",
		Level:   4,
		LogFile: "/dev/stdout",
	})
	if err != nil {
		return err
	}
	defType := reflect.TypeOf(new(plugins.Defaulter)).Elem()
	for _, cfg := range configs {
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
