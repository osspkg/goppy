package goppy

import (
	"flag"
	"os"
	"reflect"

	"github.com/dewep-online/goppy/plugins"
	"github.com/deweppro/go-app/application"
	"github.com/deweppro/go-errors"
	"gopkg.in/yaml.v3"
)

type (
	app struct {
		app    *application.App
		config string
		plgs   []interface{}
		cfgs   []interface{}
	}

	Gopper interface {
		WithConfig(filename string)
		Plugins(args ...plugins.Plugin)
		Run()
	}
)

// New constructor for init Goppy
func New() Gopper {
	return &app{
		app:  application.New(),
		plgs: make([]interface{}, 0, 100),
		cfgs: make([]interface{}, 0, 100),
	}
}

// WithConfig set config path for goppy
func (v *app) WithConfig(filename string) {
	v.config = filename
}

// Plugins setting the list of plugins to initialize
func (v *app) Plugins(args ...plugins.Plugin) {
	for _, arg := range args {
		reflectResolve(arg.Config, reflect.Ptr, func(in interface{}) {
			v.cfgs = append(v.cfgs, in)
		}, "Plugin.Config can only be a reference to an object")
		reflectResolve(arg.Inject, reflect.Func, func(in interface{}) {
			v.plgs = append(v.plgs, in)
		}, "Plugin.Inject can only be a function that accepts dependencies and returns a reference to the initialized service")
		reflectResolve(arg.Resolve, reflect.Func, func(in interface{}) {
			v.plgs = append(v.plgs, in)
		}, "Plugin.Resolve can only be a function that accepts dependencies")
	}
}

// Run launching Goppy with initialization of all dependencies
func (v *app) Run() {
	v.config = generateConfig(v.config, v.cfgs...)

	v.app.ConfigFile(v.config, v.cfgs...).
		Modules(v.plgs...).
		Run()
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

func parseConfigFlag() string {
	conf := flag.String("config", "./config.yaml", "path to the config file")
	flag.Parse()
	return *conf
}

// nolint: unparam
func generateConfig(filename string, cfgs ...interface{}) string {
	if len(filename) == 0 {
		filename = parseConfigFlag()
	}
	if _, err := os.Stat(filename); !errors.Is(err, os.ErrNotExist) {
		return filename
	}
	//nolint: errcheck
	b, _ := yaml.Marshal(&application.BaseConfig{
		Env:     "dev",
		Level:   4,
		LogFile: "/dev/stdout",
	})
	defType := reflect.TypeOf(new(plugins.Defaulter)).Elem()
	for _, cfg := range cfgs {
		if reflect.TypeOf(cfg).AssignableTo(defType) {
			reflect.ValueOf(cfg).MethodByName("Default").Call([]reflect.Value{})
		}
		if bb, err := yaml.Marshal(cfg); err == nil {
			b = append(b, '\n')
			b = append(b, bb...)
		}
	}
	if err := os.WriteFile(filename, b, 0755); err != nil {
		panic(err)
	}
	return filename
}
