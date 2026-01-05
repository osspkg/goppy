/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import (
	"go.osspkg.com/network/address"
)

type ZoneResolve struct {
	dns []string
}

func DefaultZoneResolve(dns ...string) *ZoneResolve {
	if len(dns) == 0 {
		dns = append(dns, "1.1.1.1", "1.0.0.1", "8.8.8.8", "8.8.4.4")
	}
	dns = address.FixIPPort("53", dns...)
	return &ZoneResolve{dns: dns}
}

func (v *ZoneResolve) Resolve(name string) []string {
	return append(make([]string, 0, len(v.dns)), v.dns...)
}

func DefaultExchanger(dns ...string) HandlerDNS {
	cli := NewClient(WithNetUDP())
	cli.SetZoneResolver(DefaultZoneResolve(dns...))
	return cli
}
