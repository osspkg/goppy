package xdns

import "time"

type (
	Config struct {
		DNS ConfigItem `yaml:"dns"`
	}
	ConfigItem struct {
		Addr    string        `yaml:"addr"`
		Timeout time.Duration `yaml:"timeout"`
	}
)

func (v *Config) Default() {
	v.DNS.Addr = "0.0.0.0:53"
	v.DNS.Timeout = 5 * time.Second
}
