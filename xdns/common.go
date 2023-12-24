package xdns

import (
	"github.com/miekg/dns"
)

type Exchanger interface {
	Exchange(q []dns.Question) ([]dns.RR, error)
}

type ZoneResolver interface {
	Resolve(name string) string
}
