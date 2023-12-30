/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package server

import (
	"crypto/tls"
	"fmt"
	"time"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

type Server struct {
	conf      ConfigItem
	listeners []*Listen
	handler   HandlerTCP
	log       xlog.Logger
	wg        iosync.Group
	sync      iosync.Switch
}

func New(conf ConfigItem, l xlog.Logger) *Server {
	return &Server{
		conf:      conf,
		listeners: make([]*Listen, 0, len(conf.Pools)),
		handler:   NewLogHandlerTCP(l),
		log:       l,
		wg:        iosync.NewGroup(),
		sync:      iosync.NewSwitch(),
	}
}

func (v *Server) HandleFunc(h HandlerTCP) {
	if v.sync.IsOff() {
		v.handler = h
	}
}

func (v *Server) Up(ctx xc.Context) error {
	if !v.sync.On() {
		return fmt.Errorf("server already running")
	}

	v.log.Infof("TCP server starting")

	if err := v.buildListeners(); err != nil {
		return err
	}

	v.runListeners(ctx)

	return nil
}

func (v *Server) Down() error {
	if !v.sync.Off() {
		return fmt.Errorf("server already stopped")
	}
	v.log.Infof("TCP server stopping")
	var err error
	for _, l := range v.listeners {
		err = errors.Wrap(err, l.Close())
	}
	v.wg.Wait()
	return err
}

func (v *Server) buildListeners() error {
	if len(v.conf.Pools) == 0 {
		return fmt.Errorf("settings pool is empty")
	}
	for _, c := range v.conf.Pools {
		certs := make([]Cert, 0, len(c.Certs))
		for _, cert := range c.Certs {
			certs = append(certs, Cert{Public: cert.Public, Private: cert.Public})
		}
		l, err := NewListen(c.Port, certs...)
		if err != nil {
			v.log.WithFields(xlog.Fields{
				"port": c.Port,
				"err":  err.Error(),
			}).Errorf("TCP server starting")
			return err
		}
		v.log.WithField("port", c.Port).Infof("TCP server starting")
		v.listeners = append(v.listeners, l)
	}
	return nil
}

func (v *Server) runListeners(ctx xc.Context) {
	for _, l := range v.listeners {
		l := l
		v.wg.Background(func() {
			for {
				conn, err := l.Accept()

				select {
				case <-ctx.Done():
					return
				default:
				}

				if err != nil {
					v.log.WithFields(xlog.Fields{
						"err":    err.Error(),
						"action": "accept",
					}).Errorf("TCP Handler")
					continue
				}

				if v.conf.Timeout > 0 {
					err = errors.Wrap(
						conn.SetDeadline(time.Now().Add(v.conf.Timeout)),
						conn.SetReadDeadline(time.Now().Add(v.conf.Timeout)),
						conn.SetWriteDeadline(time.Now().Add(v.conf.Timeout)),
					)
					if err != nil {
						v.log.WithFields(xlog.Fields{
							"err":    err.Error(),
							"action": "set deadline",
						}).Errorf("TCP Handler")
					}
				}

				if tc, ok := conn.(*tls.Conn); ok {
					if err = tc.HandshakeContext(ctx.Context()); err != nil {
						err = errors.Wrap(err, conn.Close())
						v.log.WithFields(xlog.Fields{
							"err":    err.Error(),
							"action": "tls handshake",
						}).Errorf("TCP Handler")
						continue
					}
				}

				v.wg.Background(func() {
					v.handler.HandlerTCP(newConnProcessor(conn, v.conf.Timeout))
					if err0 := conn.Close(); err0 != nil {
						v.log.WithFields(xlog.Fields{
							"err":    err0.Error(),
							"action": "close",
						}).Errorf("TCP Handler")
					}
				})
			}
		})
	}
}
