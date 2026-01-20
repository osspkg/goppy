/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import (
	"strings"

	"github.com/miekg/dns"
)

var (
	_qtypeMapUTS = make(map[uint16]string, len(dns.TypeToString))
	_qtypeMapSTU = make(map[string]uint16, len(dns.TypeToString))
)

func init() {
	for u, s := range dns.TypeToString {
		_qtypeMapUTS[u] = s
		_qtypeMapSTU[s] = u
	}
}

func QTypeUint16(s string) uint16 {
	s = strings.ToUpper(s)
	if v, ok := _qtypeMapSTU[s]; ok {
		return v
	}
	return 0
}

func QTypeString(u uint16) string {
	if v, ok := _qtypeMapUTS[u]; ok {
		return v
	}
	return ""
}
