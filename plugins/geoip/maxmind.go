package geoip

import (
	"fmt"
	"net"

	"github.com/dewep-online/goppy/plugins"
	"github.com/deweppro/go-logger"
	"github.com/oschwald/geoip2-golang"
)

//MaxMindConfig MaxMind database config
type MaxMindConfig struct {
	DB string `yaml:"maxminddb"`
}

func (v *MaxMindConfig) Default() {
	v.DB = "./GeoIP2-City.mmdb"
}

//WithMaxMindGeoIP information resolver through local MaxMind database
func WithMaxMindGeoIP() plugins.Plugin {
	return plugins.Plugin{
		Config: &MaxMindConfig{},
		Inject: func(conf *MaxMindConfig, log logger.Logger) (*maxmind, GeoIP) {
			mmdb := newMMDB(conf)
			return mmdb, mmdb
		},
	}
}

type (
	//GeoIP geo-ip information definition interface
	GeoIP interface {
		Country(ip net.IP) (string, error)
	}

	maxmind struct {
		conf *MaxMindConfig
		db   *geoip2.Reader
	}
)

func newMMDB(c *MaxMindConfig) *maxmind {
	return &maxmind{
		conf: c,
	}
}

func (v *maxmind) Up() error {
	db, err := geoip2.Open(v.conf.DB)
	if err != nil {
		return fmt.Errorf("maxmind: %w", err)
	}
	v.db = db
	return nil
}

func (v *maxmind) Down() error {
	if v.db != nil {
		return v.db.Close()
	}
	return nil
}

func (v *maxmind) Country(ip net.IP) (string, error) {
	vv, err := v.db.Country(ip)
	if err != nil {
		return "", err
	}
	return vv.Country.IsoCode, nil
}
