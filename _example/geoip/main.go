/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"encoding/json"
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
		plugins.Kind{
			Resolve: func(routes web.ServerPool, gip geoip.GeoIP) {
				router, ok := routes.Main()
				if !ok {
					return
				}

				router.Use(
					geoip.ResolveIPMiddleware(gip),
					geoip.HeadersMiddleware(
						geoip.HeaderCloudflareClientIP,
						geoip.HeaderCloudflareClientCountry,
					),
				)

				router.Get("/", func(ctx web.Ctx) {
					m := Model{data: struct {
						ClientIP string   `json:"client_ip"`
						Country  string   `json:"country"`
						ProxyIPs []net.IP `json:"proxy_ips"`
					}{
						ClientIP: geoip.GetClientIP(ctx.Context()).String(),
						Country:  geoip.GetCountryName(ctx.Context()),
						ProxyIPs: geoip.GetProxyIPs(ctx.Context()),
					}}
					ctx.JSON(200, &m)
				})
			},
		},
	)
	app.Run()

}

type Model struct {
	data struct {
		ClientIP string   `json:"client_ip"`
		Country  string   `json:"country"`
		ProxyIPs []net.IP `json:"proxy_ips"`
	}
}

func (m Model) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.data)
}
