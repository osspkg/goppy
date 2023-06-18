/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"net"

	"github.com/osspkg/goppy"
	"github.com/osspkg/goppy/plugins"
	"github.com/osspkg/goppy/plugins/geoip"
	"github.com/osspkg/goppy/plugins/web"
)

func main() {

	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		web.WithHTTPDebug(),
		web.WithHTTP(),
		geoip.WithMaxMindGeoIP(),
	)
	app.Plugins(
		plugins.Plugin{
			Resolve: func(routes web.RouterPool, gip geoip.GeoIP) {
				router := routes.Main()
				router.Use(
					geoip.CloudflareMiddleware(),
					geoip.MaxMindMiddleware(gip),
				)
				router.Get("/", func(ctx web.Context) {
					m := model{
						ClientIP: geoip.GetClientIP(ctx).String(),
						Country:  geoip.GetCountryName(ctx),
						ProxyIPs: geoip.GetProxyIPs(ctx),
					}
					ctx.JSON(200, &m)
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
