package goppy

import (
	"os"
	"reflect"

	"github.com/dewep-online/goppy/plugins"
	"github.com/deweppro/go-app/application"
	"github.com/deweppro/go-errors"
	"gopkg.in/yaml.v3"
)

type goppy struct {
	app    *application.App
	config string
	plgs   []interface{}
	cfgs   []interface{}
}

//New constructor for init Goppy
func New() *goppy {
	return &goppy{
		app:  application.New(),
		plgs: make([]interface{}, 0, 100),
		cfgs: make([]interface{}, 0, 100),
	}
}

//WithConfig set config path for goppy
func (v *goppy) WithConfig(filename string) {
	v.config = filename
}

func (v *goppy) resolve(arg interface{}, k reflect.Kind, call func(interface{}), comment string) {
	if arg == nil {
		return
	}
	if reflect.TypeOf(arg).Kind() != k {
		panic(comment)
	}
	call(arg)
}

//Plugins setting the list of plugins to initialize
func (v *goppy) Plugins(args ...plugins.Plugin) {
	for _, arg := range args {
		v.resolve(arg.Config, reflect.Ptr, func(in interface{}) {
			v.cfgs = append(v.cfgs, in)
		}, "Plugin.Config can only be a reference to an object")
		v.resolve(arg.Inject, reflect.Func, func(in interface{}) {
			v.plgs = append(v.plgs, in)
		}, "Plugin.Inject can only be a function that accepts dependencies and returns a reference to the initialized service")
		v.resolve(arg.Resolve, reflect.Func, func(in interface{}) {
			v.plgs = append(v.plgs, in)
		}, "Plugin.Resolve can only be a function that accepts dependencies")
	}
}

//Run launching Goppy with initialization of all dependencies
func (v *goppy) Run() {
	if len(v.config) == 0 {
		v.WithConfig(createTempConfig(v.cfgs...))
	}

	v.app.ConfigFile(v.config, v.cfgs...).
		Modules(v.plgs...).
		Run()
}

//nolint: unparam
func createTempConfig(cfgs ...interface{}) string {
	filename := "./config.yaml"
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
