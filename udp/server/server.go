/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package server

import (
	"fmt"
	"net"
	"strings"

	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
	"go.osspkg.com/goppy/xnet"
)

type (
	Server struct {
		addr    string
		conn    net.PacketConn
		handler HandlerUDP
		log     xlog.Logger
		wg      iosync.Group
		sync    iosync.Switch
	}
)

func New(l xlog.Logger, addr string) *Server {
	return &Server{
		addr:    addr,
		handler: NewLogHandlerUDP(l),
		log:     l,
		sync:    iosync.NewSwitch(),
		wg:      iosync.NewGroup(),
	}
}

func (v *Server) HandleFunc(h HandlerUDP) {
	if v.sync.IsOff() {
		v.handler = h
	}
}

func (v *Server) Up(ctx xc.Context) error {
	if !v.sync.On() {
		return fmt.Errorf("server already running")
	}
	var err error
	v.conn, err = net.ListenPacket("udp", xnet.CheckHostPort(v.addr))
	if err != nil {
		return err
	}
	v.wg.Background(func() {
		for {
			buf := getBuf()
			n, addr, err0 := v.conn.ReadFrom(buf)
			if err0 != nil {
				if !strings.Contains(err0.Error(), "use of closed network connection") {
					v.log.WithError("err", err0).Errorf("connection read error")
				}
				ctx.Close()
				return
			}
			if n == 0 {
				continue
			}
			go func() {
				v.handler.HandlerUDP(v.conn, addr, buf[:n])
				setBuf(buf)
			}()
		}
	})
	return nil
}

func (v *Server) Down() error {
	if !v.sync.Off() {
		return fmt.Errorf("server already stopped")
	}
	err := v.conn.Close()
	v.wg.Wait()
	return err
}
