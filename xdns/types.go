/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import (
	"github.com/miekg/dns"
)

type HandlerDNS interface {
	Exchange(q dns.Question) ([]dns.RR, error)
}

type ZoneResolver interface {
	Resolve(name string) []string
}
