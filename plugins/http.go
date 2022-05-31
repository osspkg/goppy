package plugins

import (
	"github.com/deweppro/go-http/pkg/routes"
	"github.com/deweppro/go-http/servers"
	"github.com/deweppro/go-http/servers/web"
	"github.com/deweppro/go-logger"
)

type HTTPConfig struct {
	Config servers.Config `yaml:"http"`
}

func WithHTTP() Plugin {
	return Plugin{
		Config: &HTTPConfig{},
		Inject: func(conf *HTTPConfig, log logger.Logger) (*web.Server, *routes.Router) {
			route := routes.NewRouter()
			return web.New(conf.Config, route, log), route
		},
		Dependencies: nil,
	}
}
