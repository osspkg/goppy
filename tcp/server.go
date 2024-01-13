/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package tcp

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/iosync"
)

type (
	Handler interface {
		HandlerTCP(c Connect)
	}
	ErrHandler interface {
		ErrHandlerTCP(c ErrConnect)
	}
	Server struct {
		conf        ConfigItem
		listener    net.Listener
		baseHandler Handler
		errHandler  ErrHandler
		wg          iosync.Group
		sync        iosync.Switch
	}
)

func NewServer(conf ConfigItem) *Server {
	return &Server{
		conf: conf,
		wg:   iosync.NewGroup(),
		sync: iosync.NewSwitch(),
	}
}

func (v *Server) HandleFunc(h Handler) {
	if v.sync.IsOn() {
		return
	}
	v.baseHandler = h
}

func (v *Server) ErrHandleFunc(h ErrHandler) {
	if v.sync.IsOn() {
		return
	}
	v.errHandler = h
}

func (v *Server) ListenAndServe(ctx context.Context) error {
	if v.baseHandler == nil {
		return fmt.Errorf("handler not found")
	}
	if v.errHandler == nil {
		return fmt.Errorf("error handler not found")
	}
	if !v.sync.On() {
		return errServAlreadyRunning
	}
	defer v.sync.Off()
	if err := v.build(); err != nil {
		return err
	}
	v.run(ctx)
	v.wg.Wait()
	return nil
}

func (v *Server) Close() error {
	if !v.sync.Off() {
		return nil
	}
	return v.listener.Close()
}

func (v *Server) build() error {
	certs := make([]Cert, 0, 1)
	for _, cert := range v.conf.Certs {
		certs = append(certs, Cert{Cert: cert.Cert, Key: cert.Key})
	}
	l, err := NewListen(v.conf.Address, certs...)
	if err != nil {
		return err
	}
	v.listener = l
	return nil
}

func (v *Server) run(ctx context.Context) {
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

			if tc, ok := conn.(*tls.Conn); ok {
				if err = tc.HandshakeContext(ctx); err != nil {
					fmt.Println(err)
					conn.Close() //nolint: errcheck
					continue
				}
			}

			v.wg.Background(func() {
				defer conn.Close() //nolint: errcheck
				cp := newConnectProvider(ctx, conn, cpConfig{
					MaxSize: v.conf.ClientMaxBodySize,
					Timeout: v.conf.Timeout,
				})
				defer cp.Close() //nolint: errcheck
				for {
					if err = cp.Wait(); err != nil {
						if !errors.Is(err, io.EOF) {
							v.errHandler.ErrHandlerTCP(cp)
						}
						return
					}
					if cp.IsEmpty() {
						return
					}
					v.baseHandler.HandlerTCP(cp)
				}
			})
		}
	})
}
