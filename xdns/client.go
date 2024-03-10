/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/iosync"
)

type (
	Client struct {
		cli      *dns.Client
		resolver ZoneResolver
		mux      iosync.Lock
	}

	Option func(*dns.Client)
)

func OptionNetTCP() Option {
	return func(client *dns.Client) {
		client.Net = "tcp"
	}
}

func OptionNetUDP() Option {
	return func(client *dns.Client) {
		client.Net = "udp"
	}
}

func OptionNetDOT() Option {
	return func(client *dns.Client) {
		client.Net = "tcp-tls"
	}
}

func NewClient(opts ...Option) *Client {
	cli := &Client{
		cli: &dns.Client{
			Net:          "udp",
			ReadTimeout:  time.Second * 5,
			WriteTimeout: time.Second * 5,
		},
		mux: iosync.NewLock(),
	}

	for _, opt := range opts {
		opt(cli.cli)
	}

	return cli
}

func (v *Client) SetZoneResolver(r ZoneResolver) {
	v.resolver = r
}

func (v *Client) Exchange(q []dns.Question) ([]dns.RR, error) {
	var errs error
	result := make([]dns.RR, 0, len(q))

	for _, question := range q {
		address := v.resolver.Resolve(question.Name)
		msg := new(dns.Msg).SetQuestion(question.Name, question.Qtype)

		resp, _, err := v.cli.Exchange(msg, address)
		if err != nil {
			errs = errors.Wrap(errs, errors.Wrapf(err, "name: %s, dns: %s", question.String(), address))
			continue
		}

		if len(resp.Answer) == 0 {
			errs = errors.Wrap(errs,
				errors.Wrapf(fmt.Errorf("empty answer"), "name: %s, dns: %s", question.String(), address))
			continue
		}

		result = append(result, resp.Answer...)
	}

	return result, errs
}
