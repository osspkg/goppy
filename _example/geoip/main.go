/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"net"

	"go.osspkg.com/goppy/v2"
	"go.osspkg.com/goppy/v2/geoip"
	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/web"
)

func main() {

	app := goppy.New("", "", "")
	app.Plugins(
		web.WithServer(),
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
						ClientIP: geoip.GetClientIP(ctx.Context()).String(),
						Country:  geoip.GetCountryName(ctx.Context()),
						ProxyIPs: geoip.GetProxyIPs(ctx.Context()),
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
	Country  string   `json:"country"`
	ProxyIPs []net.IP `json:"proxy_ips"`
}
