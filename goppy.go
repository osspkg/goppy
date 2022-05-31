package goppy

import (
	"github.com/dewep-online/goppy/plugins"
	"github.com/deweppro/go-app/application"
	"github.com/deweppro/go-errors"
	"gopkg.in/yaml.v3"
	"os"
)

type main struct {
	app    *application.App
	config string
	plgs   []interface{}
	cfgs   []interface{}
}

func New() *main {
	return &main{
		app:  application.New(),
		plgs: make([]interface{}, 0, 100),
		cfgs: make([]interface{}, 0, 100),
	}
}

func (v *main) WithConfig(filename string) {
	v.config = filename
}

func (v *main) Plugins(args ...plugins.Plugin) {
	for _, arg := range args {
		if arg.Config != nil {
			v.cfgs = append(v.cfgs, arg.Config)
		}
		if arg.Inject != nil {
			v.plgs = append(v.plgs, arg.Inject)
		}
		if arg.Dependencies != nil {
			v.Plugins(arg.Dependencies...)
		}
	}
}

func (v *main) Run() {
	if len(v.config) == 0 {
		v.WithConfig(createTempConfig(v.cfgs...))
	}

	v.app.ConfigFile(v.config, v.cfgs...).
		Modules(v.plgs...).
		Run()
}

func createTempConfig(cfgs ...interface{}) string {
	filename := "./config.yaml"
	if _, err := os.Stat(filename); !errors.Is(err, os.ErrNotExist) {
		return filename
	}
	b, _ := yaml.Marshal(&application.BaseConfig{
		Env:     "dev",
		Level:   4,
		LogFile: "/dev/stdout",
	})
	for _, cfg := range cfgs {
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
