/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"context"
	"net"
	"net/http"
	"time"

	"go.osspkg.com/errors"
	"go.osspkg.com/logx"
	"go.osspkg.com/network/address"
	"go.osspkg.com/syncing"
	"go.osspkg.com/xc"
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
		log     logx.Logger
		wg      syncing.Group
		sync    syncing.Switch
	}
)

// NewServer create default http server
func NewServer(name string, conf Config, handler http.Handler, l logx.Logger) *Server {
	srv := &Server{
		name:    name,
		conf:    conf,
		handler: handler,
		log:     l,
		sync:    syncing.NewSwitch(),
		wg:      syncing.NewGroup(),
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
	s.conf.Addr = address.CheckHostPort(s.conf.Addr)
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

	s.log.Info("Http server started", "name", s.name, "ip", s.conf.Addr)

	s.wg.Background(func() {
		defer ctx.Close()
		if err = s.serv.Serve(nl); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Error("Http server stopped", "name", s.name, "ip", s.conf.Addr, "err", err)
			return
		}

		s.log.Info("Http server stopped", "name", s.name, "ip", s.conf.Addr)
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
