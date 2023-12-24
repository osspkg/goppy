package xdns

import "go.osspkg.com/goppy/plugins"

func WithDNSClient(opts ...ClientOption) plugins.Plugin {
	return plugins.Plugin{
		Inject: func() *Client {
			return NewClient(opts...)
		},
	}
}
