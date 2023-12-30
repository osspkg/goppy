/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package tcp

import (
	"context"
	"crypto/tls"
	"time"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/iosync"
)

type ServerTCP struct {
	conf     ConfigItem
	listener *Listen
	handler  HandlerTCP
	wg       iosync.Group
	mux      iosync.Lock
	sync     iosync.Switch
}

func NewServerTCP(conf ConfigItem) *ServerTCP {
	return &ServerTCP{
		conf: conf,
		wg:   iosync.NewGroup(),
		mux:  iosync.NewLock(),
		sync: iosync.NewSwitch(),
	}
}

func (v *ServerTCP) HandleFunc(h HandlerTCP) {
	v.mux.Lock(func() {
		v.handler = h
	})
}

func (v *ServerTCP) ListenAndServe(ctx context.Context) error {
	if err := v.build(); err != nil {
		return err
	}
	v.sync.On()
	v.run(ctx)
	v.wg.Wait()
	return nil
}

func (v *ServerTCP) Close() error {
	if !v.sync.Off() {
		return nil
	}
	return v.listener.Close()
}

func (v *ServerTCP) build() error {
	certs := make([]Cert, 0, 1)
	for _, cert := range v.conf.Certs {
		certs = append(certs, Cert{Public: cert.Public, Private: cert.Public})
	}
	l, err := NewListen(v.conf.Address, certs...)
	if err != nil {
		return err
	}
	v.listener = l
	return nil
}

func (v *ServerTCP) run(ctx context.Context) {
	v.wg.Background(func() {
		for {
			conn, err := v.listener.Accept()

			select {
			case <-ctx.Done():
				return
			default:
			}

			if err != nil {
				continue
			}

			if v.conf.Timeout > 0 {
				err = errors.Wrap(
					conn.SetDeadline(time.Now().Add(v.conf.Timeout)),
					conn.SetReadDeadline(time.Now().Add(v.conf.Timeout)),
					conn.SetWriteDeadline(time.Now().Add(v.conf.Timeout)),
				)
				if err != nil {
					continue
				}
			}

			if tc, ok := conn.(*tls.Conn); ok {
				if err = tc.HandshakeContext(ctx); err != nil {
					conn.Close() //nolint: errcheck
					continue
				}
			}

			v.wg.Background(func() {
				defer conn.Close() //nolint: errcheck
				nc := newConnect(conn, v.conf.Timeout)
				v.mux.RLock(func() {
					if v.handler != nil {
						v.handler.HandlerTCP(nc, nc)
					}
				})
			})
		}
	})
}
