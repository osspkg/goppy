/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
	"go.osspkg.com/goppy/xnet"
)

type (
	Config struct {
		Addr            string        `yaml:"addr"`
		Network         string        `yaml:"network,omitempty"`
		ReadTimeout     time.Duration `yaml:"read_timeout,omitempty"`
		WriteTimeout    time.Duration `yaml:"write_timeout,omitempty"`
		IdleTimeout     time.Duration `yaml:"idle_timeout,omitempty"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout,omitempty"`
	}

	Server struct {
		name    string
		conf    Config
		serv    *http.Server
		handler http.Handler
		log     xlog.Logger
		wg      iosync.Group
		sync    iosync.Switch
	}
)

// NewServer create default http server
func NewServer(name string, conf Config, handler http.Handler, l xlog.Logger) *Server {
	srv := &Server{
		name:    strings.ToUpper(name),
		conf:    conf,
		handler: handler,
		log:     l,
		sync:    iosync.NewSwitch(),
		wg:      iosync.NewGroup(),
	}
	srv.validate()
	return srv
}

func (s *Server) validate() {
	if s.conf.ReadTimeout == 0 {
		s.conf.ReadTimeout = defaultTimeout
	}
	if s.conf.WriteTimeout == 0 {
		s.conf.WriteTimeout = defaultTimeout
	}
	if s.conf.IdleTimeout == 0 {
		s.conf.IdleTimeout = defaultTimeout
	}
	if s.conf.ShutdownTimeout == 0 {
		s.conf.ShutdownTimeout = defaultShutdownTimeout
	}
	if len(s.conf.Network) == 0 {
		s.conf.Network = defaultNetwork
	}
	if _, ok := networkType[s.conf.Network]; !ok {
		s.conf.Network = defaultNetwork
	}
	s.conf.Addr = xnet.CheckHostPort(s.conf.Addr)
}

// Up start http server
func (s *Server) Up(ctx xc.Context) error {
	if !s.sync.On() {
		return errors.Wrapf(errServAlreadyRunning, "starting server on %s", s.conf.Addr)
	}
	s.serv = &http.Server{
		ReadTimeout:  s.conf.ReadTimeout,
		WriteTimeout: s.conf.WriteTimeout,
		IdleTimeout:  s.conf.IdleTimeout,
		Handler:      s.handler,
	}

	nl, err := net.Listen(s.conf.Network, s.conf.Addr)
	if err != nil {
		return err
	}

	s.log.WithFields(xlog.Fields{
		"ip": s.conf.Addr,
	}).Infof("%s server started", s.name)

	s.wg.Background(func() {
		if err = s.serv.Serve(nl); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.WithFields(xlog.Fields{
				"err": err.Error(), "ip": s.conf.Addr,
			}).Errorf("%s server stopped", s.name)
			ctx.Close()
			return
		}
		s.log.WithFields(xlog.Fields{
			"ip": s.conf.Addr,
		}).Infof("%s server stopped", s.name)
	})
	return nil
}

// Down stop http server
func (s *Server) Down() error {
	if !s.sync.Off() {
		return errors.Wrapf(errServAlreadyStopped, "stopping server on %s", s.conf.Addr)
	}
	ctx, cncl := context.WithTimeout(context.Background(), s.conf.ShutdownTimeout)
	defer cncl()
	err := s.serv.Shutdown(ctx)
	s.wg.Wait()
	return err
}
