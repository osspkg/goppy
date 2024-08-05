/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import (
	"errors"
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
	log     logx.Logger
	wg      syncing.Group
	mux     syncing.Lock
}

func NewServer(conf Config, l logx.Logger) *Server {
	return &Server{
		conf:    conf,
		serv:    make([]*dns.Server, 0, 2),
		handler: DefaultExchanger(),
		wg:      syncing.NewGroup(),
		mux:     syncing.NewLock(),
		log:     l,
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
			v.log.Info("Start DNS Server", "address", srv.Addr, "net", srv.Net)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				v.log.Error("Start DNS Server", "address", srv.Addr, "net", srv.Net, "err", err)
				ctx.Close()
			}
		})
	}

	return nil
}

func (v *Server) Down() error {
	for _, srv := range v.serv {
		if err := srv.Shutdown(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			v.log.Error("Shutdown DNS Server", "address", srv.Addr, "net", srv.Net, "err", err)
			continue
		}
		v.log.Info("Shutdown DNS Server", "address", srv.Addr, "net", srv.Net)
	}

	v.wg.Wait()
	return nil
}

func (v *Server) HandleFunc(r HandlerDNS) {
	v.mux.Lock(func() {
		v.handler = r
	})
}

func (v *Server) dnsHandler(w dns.ResponseWriter, msg *dns.Msg) {
	response := &dns.Msg{}
	response.Authoritative = true
	response.RecursionAvailable = true
	response.SetReply(msg)
	response.SetRcode(msg, dns.RcodeSuccess)

	var (
		answer []dns.RR
		err    error
	)
	v.mux.RLock(func() {
		answer, err = v.handler.Exchange(msg.Question)
	})

	if err != nil {
		v.log.Error("DNS handler", "question", msg, "err", err)
	} else {
		response.Answer = append(response.Answer, answer...)
	}

	if err = w.WriteMsg(response); err != nil {
		v.log.Error("DNS handler", "question", msg, "answer", response, "err", err)
	}
}
