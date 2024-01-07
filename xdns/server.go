/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import (
	"errors"
	"net/http"

	"github.com/miekg/dns"
	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

type Server struct {
	conf    ConfigItem
	serv    []*dns.Server
	handler HandlerDNS
	log     xlog.Logger
	wg      iosync.Group
	mux     iosync.Lock
}

func NewServer(conf ConfigItem, l xlog.Logger) *Server {
	return &Server{
		conf:    conf,
		serv:    make([]*dns.Server, 0, 2),
		handler: DefaultExchanger(),
		wg:      iosync.NewGroup(),
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

	for _, s := range v.serv {
		s := s
		v.wg.Background(func() {
			v.log.WithFields(xlog.Fields{
				"address": s.Addr,
				"net":     s.Net,
			}).Infof("Run DNS Server")
			if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				v.log.WithFields(xlog.Fields{
					"err":     err.Error(),
					"address": s.Addr,
					"net":     s.Net,
				}).Errorf("Run DNS Server")
				ctx.Close()
			}
		})
	}

	return nil
}

func (v *Server) Down() error {
	for _, s := range v.serv {
		if err := s.Shutdown(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			v.log.WithFields(xlog.Fields{
				"err":     err.Error(),
				"address": s.Addr,
				"net":     s.Net,
			}).Errorf("Shutdown DNS Server")
			continue
		}
		v.log.WithFields(xlog.Fields{
			"address": s.Addr,
			"net":     s.Net,
		}).Infof("Shutdown DNS Server")
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
		v.log.WithFields(xlog.Fields{
			"err":      err.Error(),
			"question": msg.String(),
		}).Errorf("DNS handler")
	} else {
		response.Answer = append(response.Answer, answer...)
	}

	if err = w.WriteMsg(response); err != nil {
		v.log.WithFields(xlog.Fields{
			"err":      err.Error(),
			"question": msg.String(),
			"answer":   response.String(),
		}).Errorf("DNS handler")
	}
}
