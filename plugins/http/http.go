package http

import (
	"github.com/dewep-online/goppy/plugins"
	"github.com/deweppro/go-http/servers"
	"github.com/deweppro/go-logger"
)

//Config config to initialize HTTP service
type Config struct {
	Config map[string]servers.Config `yaml:"http"`
}

func (v *Config) Default() {
	if v.Config == nil {
		v.Config = map[string]servers.Config{
			"main": {Addr: "127.0.0.1:8080"},
		}
	}
}

//WithHTTP launch of HTTP service with default Router
func WithHTTP() plugins.Plugin {
	return plugins.Plugin{
		Config: &Config{},
		Inject: func(conf *Config, log logger.Logger) (*routeProvider, RouterPool) {
			rp := newRouteProvider(conf.Config, log)
			return rp, rp
		},
	}
}
