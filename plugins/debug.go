package plugins

import (
	"github.com/deweppro/go-http/servers"
	"github.com/deweppro/go-http/servers/debug"
	"github.com/deweppro/go-logger"
)

type DebugConfig struct {
	Config servers.Config `yaml:"debug"`
}

func WithHTTPDebug() Plugin {
	return Plugin{
		Config: &DebugConfig{},
		Inject: func(conf *DebugConfig, log logger.Logger) *debug.Debug {
			return debug.New(conf.Config, log)
		},
		Dependencies: nil,
	}
}
