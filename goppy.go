/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package goppy

import (
	"fmt"
	"os"

	"go.osspkg.com/config"
	configEnv "go.osspkg.com/config/env"
	"go.osspkg.com/console"
	"go.osspkg.com/errors"
	"go.osspkg.com/goppy/v2/env"
	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/grape"
	grapeConfig "go.osspkg.com/grape/config"
	"go.osspkg.com/ioutils/codec"
	"go.osspkg.com/logx"
)

type (
	_config struct {
		Filename string
		Data     string
		Ext      string
	}

	_app struct {
		info  *env.AppInfo
		grape grape.Grape

		commands map[string]interface{}

		plugins []interface{}
		configs []interface{}

		cfg _config

		resolvers []config.Resolver
		args      *console.Args
	}

	Goppy interface {
		Logger(l logx.Logger)
		Plugins(args ...plugins.Plugin)
		Command(name string, call interface{})
		ConfigResolvers(rc ...config.Resolver)
		ConfigData(data, ext string)
		Run()
	}
)

// New constructor for init Goppy
func New(name, version, description string) Goppy {
	return &_app{
		grape: grape.New(name).ExitFunc(func(code int) {
			os.Exit(code)
		}),
		commands:  make(map[string]interface{}),
		plugins:   make([]interface{}, 0, 100),
		configs:   make([]interface{}, 0, 100),
		resolvers: make([]config.Resolver, 0, 100),
		args:      console.NewArgs().Parse(os.Args[1:]),
		info: func() *env.AppInfo {
			info := env.NewAppInfo()
			info.AppName = env.AppName(name)
			info.AppVersion = env.AppVersion(version)
			info.AppDescription = env.AppDescription(description)
			return &info
		}(),
	}
}

func (v *_app) Logger(l logx.Logger) {
	v.grape.Logger(l)
}

func (v *_app) ConfigResolvers(rc ...config.Resolver) {
	v.resolvers = append(v.resolvers, rc...)
}

func (v *_app) ConfigData(data, ext string) {
	v.cfg = _config{
		Filename: "",
		Data:     data,
		Ext:      ext,
	}
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
		v.ConfigResolvers(configEnv.New())
	}

	apps := v.grape.Modules(v.plugins...)
	apps.Modules(v.info.AppName, v.info.AppVersion, v.info.AppDescription, *v.info)

	if len(v.cfg.Data) > 0 {
		apps.ConfigData(v.cfg.Data, v.cfg.Ext)
	} else {
		v.cfg.Filename = v.parseConfigFlag()
		console.FatalIfErr(v.recoveryConfig(v.cfg.Filename), "config recovery")
		apps.ConfigFile(v.cfg.Filename)
	}
	console.FatalIfErr(v.validateConfig(v.cfg), "config validate")

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

func (v *_app) validateConfig(c _config) error {
	rc := config.New(v.resolvers...)

	switch true {
	case len(c.Filename) > 0:
		if err := rc.OpenFile(c.Filename); err != nil {
			return err
		}
	case len(c.Data) > 0:
		rc.OpenBlob(c.Data, c.Ext)
	default:
		return nil
	}

	if err := rc.Build(); err != nil {
		return err
	}

	for _, cfg := range v.configs {
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

	for _, cfg := range v.configs {
		if vv, ok := cfg.(plugins.Defaulter); ok {
			vv.Default()
		}
	}

	cfg := &grapeConfig.Config{
		Env: "dev",
		Log: grapeConfig.LogConfig{
			Level:    4,
			FilePath: "/dev/stdout",
			Format:   "string",
		},
	}

	return codec.FileEncoder(filename).Encode(append(v.configs, cfg)...)
}
