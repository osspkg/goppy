package goppy

import (
	"flag"
	"os"
	"reflect"

	"github.com/deweppro/go-sdk/app"
	"github.com/deweppro/go-sdk/console"
	"github.com/deweppro/go-sdk/errors"
	"github.com/deweppro/goppy/plugins"
	"gopkg.in/yaml.v3"
)

type (
	_app struct {
		application app.App
		console     *console.Console
		config      string
		plugins     []interface{}
		configs     []interface{}
	}

	Goppy interface {
		WithConfig(filename string)
		Plugins(args ...plugins.Plugin)
		Command(call func(s console.CommandSetter))
		Run()
	}
)

// New constructor for init Goppy
func New() Goppy {
	return &_app{
		application: app.New(),
		console:     console.New("<app>", ""),
		plugins:     make([]interface{}, 0, 100),
		configs:     make([]interface{}, 0, 100),
	}
}

// WithConfig set config path for goppy
func (v *_app) WithConfig(filename string) {
	v.config = filename
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

func (v *_app) Command(call func(s console.CommandSetter)) {
	newCmd := console.NewCommand(call)
	v.console.AddCommand(newCmd)
}

// Run launching Goppy with initialization of all dependencies
func (v *_app) Run() {
	v.config = generateConfig(v.config, v.configs...)

	v.console.RootCommand(console.NewCommand(func(s console.CommandSetter) {
		s.Setup("root", "run app as service")
		s.ExecFunc(func(_ []string) {
			v.application.
				ConfigFile(v.config, v.configs...).
				Modules(v.plugins...).
				Run()
		})
	}))

	v.console.Exec()
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

func generateConfig(filename string, configs ...interface{}) string {
	if len(filename) == 0 {
		filename = parseConfigFlag()
	}
	if _, err := os.Stat(filename); !errors.Is(err, os.ErrNotExist) {
		return filename
	}
	//nolint: errcheck
	b, _ := yaml.Marshal(&app.Config{
		Env:     "dev",
		Level:   4,
		LogFile: "/dev/stdout",
	})
	defType := reflect.TypeOf(new(plugins.Defaulter)).Elem()
	for _, cfg := range configs {
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
