package http

import (
	"github.com/dewep-online/goppy/plugins"
	"github.com/deweppro/go-http/servers"
	"github.com/deweppro/go-http/servers/debug"
	"github.com/deweppro/go-logger"
)

//DebugConfig config to initialize HTTP debug service
type DebugConfig struct {
	Config servers.Config `yaml:"debug"`
}

func (v *DebugConfig) Default() {
	v.Config = servers.Config{Addr: "127.0.0.1:12000"}
}

//WithHTTPDebug debug service over HTTP protocol with pprof enabled
func WithHTTPDebug() plugins.Plugin {
	return plugins.Plugin{
		Config: &DebugConfig{},
		Inject: func(conf *DebugConfig, log logger.Logger) *debug.Debug {
			return debug.New(conf.Config, log)
		},
	}
}
