/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"go.osspkg.com/errors"
	"go.osspkg.com/logx"
	"go.osspkg.com/network/address"
	"go.osspkg.com/network/listen"
	"go.osspkg.com/syncing"
	"go.osspkg.com/xc"
)

type (
	Server struct {
		conf    Config
		serv    *http.Server
		handler http.Handler
		wg      syncing.Group
		sync    syncing.Switch
	}
)

// NewServer create default http server
func NewServer(ctx context.Context, conf Config, handler http.Handler) *Server {
	srv := &Server{
		conf:    conf,
		handler: handler,
		sync:    syncing.NewSwitch(),
		wg:      syncing.NewGroup(ctx),
	}
	srv.validate()
	return srv
}

func (v *Server) validate() {
	if v.conf.ReadTimeout == 0 {
		v.conf.ReadTimeout = defaultTimeout
	}
	if v.conf.WriteTimeout == 0 {
		v.conf.WriteTimeout = defaultTimeout
	}
	if v.conf.IdleTimeout == 0 {
		v.conf.IdleTimeout = defaultTimeout
	}
	if v.conf.ShutdownTimeout == 0 {
		v.conf.ShutdownTimeout = defaultShutdownTimeout
	}
	if len(v.conf.Network) == 0 {
		v.conf.Network = defaultNetwork
	}
	if _, ok := networkType[v.conf.Network]; !ok {
		v.conf.Network = defaultNetwork
	}
	v.conf.Addr = address.ResolveIPPort(v.conf.Addr)
}

func (v *Server) Up(ctx xc.Context) error {
	if !v.sync.On() {
		return errors.Wrapf(errServAlreadyRunning, "http server: starting server on '%s'", v.conf.Addr)
	}

	v.serv = &http.Server{
		ReadTimeout:  v.conf.ReadTimeout,
		WriteTimeout: v.conf.WriteTimeout,
		IdleTimeout:  v.conf.IdleTimeout,
		Handler:      v.handler,
	}

	ln, err := listen.New(ctx.Context(), v.conf.Network, v.conf.Addr, &listen.SSL{Certs: v.conf.Tls})
	if err != nil {
		return err
	}

	nl, ok := ln.(net.Listener)
	if !ok {
		return fmt.Errorf("http server: does not implement net.Listener for tag '%s'", v.conf.Tag)
	}

	v.wg.Background("http server", func(_ context.Context) {
		defer ctx.Close()

		logx.Info("HTTP Server",
			"do", "start",
			"tag", v.conf.Tag,
			"net", v.conf.Network,
			"ip", v.conf.Addr)

		servErr := v.serv.Serve(nl)

		logx.Warn("HTTP Server",
			"do", "stop",
			"err", servErr,
			"tag", v.conf.Tag,
			"net", v.conf.Network,
			"ip", v.conf.Addr)
	})
	return nil
}

func (v *Server) Down() error {
	if !v.sync.Off() {
		return errors.Wrapf(errServAlreadyStopped, "http server: stopping server on %s", v.conf.Addr)
	}

	defer v.wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), v.conf.ShutdownTimeout)
	defer cancel()

	return v.serv.Shutdown(ctx)
}
