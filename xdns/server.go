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
	wg      syncing.Group
}

func NewServer(conf Config) *Server {
	return &Server{
		conf:    conf,
		serv:    make([]*dns.Server, 0, 2),
		handler: DefaultExchanger(),
		wg:      syncing.NewGroup(),
	}
}

func (v *Server) Up(ctx xc.Context) error {
	handler := dns.NewServeMux()
	handler.HandleFunc(".", v.dnsHandler)

	v.serv = append(v.serv, &dns.Server{
		Addr:         v.conf.Addr,
		Net:          "tcp",
		Handler:      handler,
		ReadTimeout:  v.conf.Timeout,
		WriteTimeout: v.conf.Timeout,
	})
	v.serv = append(v.serv, &dns.Server{
		Addr:         v.conf.Addr,
		Net:          "udp",
		Handler:      handler,
		UDPSize:      65535,
		ReadTimeout:  v.conf.Timeout,
		WriteTimeout: v.conf.Timeout,
	})

	for _, srv := range v.serv {
		srv := srv
		v.wg.Background(func() {
			logx.Info("Start DNS Server", "address", srv.Addr, "net", srv.Net)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logx.Error("Start DNS Server", "address", srv.Addr, "net", srv.Net, "err", err)
				ctx.Close()
			}
		})
	}

	return nil
}

func (v *Server) Down() error {
	for _, srv := range v.serv {
		if err := srv.Shutdown(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logx.Error("Shutdown DNS Server", "address", srv.Addr, "net", srv.Net, "err", err)
			continue
		}
		logx.Info("Shutdown DNS Server", "address", srv.Addr, "net", srv.Net)
	}

	v.wg.Wait()
	return nil
}

func (v *Server) HandleFunc(r HandlerDNS) {
	v.handler = r
}

func (v *Server) dnsHandler(w dns.ResponseWriter, msg *dns.Msg) {
	response := &dns.Msg{}
	response.Authoritative = true
	response.RecursionAvailable = true

	for _, q := range msg.Question {
		answer, err := v.handler.Exchange(q)
		if err != nil {
			logx.Error("DNS exchange", "question", msg, "err", err)
		} else {
			response.Answer = append(response.Answer, answer...)
		}
	}

	if len(response.Answer) == 0 {
		response.SetRcode(msg, dns.RcodeNotZone)
	} else {
		response.SetRcode(msg, dns.RcodeSuccess)
	}

	fmt.Println(response.String())

	if err := w.WriteMsg(response); err != nil {
		logx.Error("DNS response", "question", msg, "answer", response, "err", err)
	}
}
