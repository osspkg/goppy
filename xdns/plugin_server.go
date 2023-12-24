package xdns

import (
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/xlog"
)

func WithDNSServer() plugins.Plugin {
	return plugins.Plugin{
		Config: &Config{},
		Inject: func(c *Config, l xlog.Logger) *Server {
			return NewServer(c.DNS, l)
		},
	}
}
