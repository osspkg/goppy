package main

import (
	"net"

	"github.com/dewep-online/goppy/middlewares"

	"github.com/dewep-online/goppy/plugins/geoip"

	"github.com/dewep-online/goppy"
	"github.com/dewep-online/goppy/plugins"
	"github.com/dewep-online/goppy/plugins/http"
)

func main() {

	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		http.WithHTTPDebug(),
		http.WithHTTP(),
		geoip.WithMaxMindGeoIP(),
	)
	app.Plugins(
		plugins.Plugin{
			Resolve: func(routes http.RouterPool, gip geoip.GeoIP) {
				router := routes.Main()
				router.Use(
					middlewares.CloudflareMiddleware(),
					middlewares.MaxMindMiddleware(gip),
				)
				router.Get("/", func(ctx http.Ctx) {
					m := model{
						ClientIP: geoip.GetClientIP(ctx).String(),
						Country:  geoip.GetCountryName(ctx),
						ProxyIPs: geoip.GetProxyIPs(ctx),
					}
					ctx.SetBody().JSON(&m)
				})
			},
		},
	)
	app.Run()

}

type model struct {
	ClientIP string   `json:"client_ip"`
	Country  *string  `json:"country"`
	ProxyIPs []net.IP `json:"proxy_ips"`
}
