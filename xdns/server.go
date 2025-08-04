/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/miekg/dns"
	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"
	"go.osspkg.com/xc"
)

type Server struct {
	conf    Config
	serv    []*dns.Server
	handler HandlerDNS
	qtypes  map[uint16]struct{}
	wg      syncing.Group
}

func NewServer(conf Config) *Server {
	return &Server{
		conf:    conf,
		serv:    make([]*dns.Server, 0, 2),
		qtypes:  make(map[uint16]struct{}, len(conf.QTypes)),
		handler: DefaultExchanger(),
		wg:      syncing.NewGroup(),
	}
}

func (v *Server) Up(ctx xc.Context) error {
	if len(v.conf.QTypes) == 0 {
		for val := range _qtypeMapUTS {
			v.qtypes[val] = struct{}{}
		}
	} else {
		for _, qt := range v.conf.QTypes {
			if val := QTypeUint16(qt); val > 0 {
				v.qtypes[val] = struct{}{}
			}
		}
	}

	handler := dns.NewServeMux()
	handler.HandleFunc(".", v.dnsHandler)

	v.serv = append(v.serv, &dns.Server{
		Addr:    v.conf.Addr,
		Net:     "tcp",
		Handler: handler,
	})
	v.serv = append(v.serv, &dns.Server{
		Addr:    v.conf.Addr,
		Net:     "udp",
		Handler: handler,
	})

	for _, srv := range v.serv {
		srv := srv
		v.wg.Background(func() {
			defer ctx.Close()
			logx.Info("DNS Server", "do", "start", "address", srv.Addr, "net", srv.Net)

			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logx.Error("DNS Server", "do", "start", "address", srv.Addr, "net", srv.Net, "err", err)
			}
		})
	}

	return nil
}

func (v *Server) Down() error {
	for _, srv := range v.serv {
		if err := srv.Shutdown(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logx.Error("DNS Server", "do", "stop", "address", srv.Addr, "net", srv.Net, "err", err)
		} else {
			logx.Info("DNS Server", "do", "stop", "address", srv.Addr, "net", srv.Net)
		}
	}

	v.wg.Wait()
	return nil
}

func (v *Server) HandleFunc(r HandlerDNS) {
	v.handler = r
}

func (v *Server) dnsHandler(w dns.ResponseWriter, msg *dns.Msg) {
	defer func() {
		if err := recover(); err != nil {
			logx.Error("DNS Server", "do", "handler: panic", "question", msg, "err", fmt.Errorf("%+v", err))
		}
		if err := w.Close(); err != nil {
			logx.Error("DNS Server", "do", "handler: close connect", "question", msg, "err", err)
		}
	}()

	response := new(dns.Msg)
	response.Authoritative = true
	response.RecursionAvailable = true
	response.SetRcode(msg, dns.RcodeSuccess)

	for _, q := range msg.Question {
		if _, ok := v.qtypes[q.Qtype]; !ok {
			continue
		}

		answers, err := v.handler.Exchange(q)
		if err != nil {
			logx.Error("DNS Server", "do", "exchange", "domain", q.Name, "qtype", QTypeString(q.Qtype), "err", err)
		} else {
			for _, answer := range answers {
				if answer == nil {
					continue
				}
				response.Answer = append(response.Answer, answer)
			}
		}
	}

	if err := w.WriteMsg(response); err != nil {
		logx.Error("DNS Server", "do", "response: write", "question", msg, "answer", response, "err", err)
	}
}
