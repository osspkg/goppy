/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import (
	"time"

	"github.com/miekg/dns"
	"go.osspkg.com/errors"
	"go.osspkg.com/syncing"
)

type (
	Client struct {
		cli      *dns.Client
		resolver ZoneResolver
		mux      syncing.Lock
	}

	Option func(*dns.Client)
)

func WithNetTCP() Option {
	return func(client *dns.Client) {
		client.Net = "tcp"
	}
}

func WithNetUDP() Option {
	return func(client *dns.Client) {
		client.Net = "udp"
	}
}

func WithNetDOT() Option {
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
		mux: syncing.NewLock(),
	}

	for _, opt := range opts {
		opt(cli.cli)
	}

	return cli
}

func (v *Client) SetZoneResolver(r ZoneResolver) {
	v.resolver = r
}

func (v *Client) Exchange(question dns.Question) ([]dns.RR, error) {
	var errs error
	msg := new(dns.Msg).SetQuestion(question.Name, question.Qtype)
	ns := v.resolver.Resolve(question.Name)

	for _, address := range ns {
		resp, _, err := v.cli.Exchange(msg, address)
		if err != nil {
			errs = errors.Wrap(errs, errors.Wrapf(err, "dns client: name: %s, dns: %s", question.Name, address))
			continue
		}

		if len(resp.Answer) == 0 {
			continue
		}

		return resp.Answer, nil
	}

	return nil, errs
}
