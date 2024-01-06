/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package client

import (
	"net"
	"strings"

	"go.osspkg.com/goppy/iosync"
)

type Client struct {
	addr string
	conn *net.UDPConn
	wg   iosync.Group
	call HandlerUDP
	mux  iosync.Lock
}

func New(addr string) (*Client, error) {
	ua, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, ua)
	if err != nil {
		return nil, err
	}
	c := &Client{
		addr: addr,
		conn: conn,
		call: newNilHandler(),
		wg:   iosync.NewGroup(),
		mux:  iosync.NewLock(),
	}
	c.readLoop()
	return c, nil
}

func (v *Client) Close() error {
	err := v.conn.Close()
	v.wg.Wait()
	return err
}

func (v *Client) Write(b []byte) (int, error) {
	return v.conn.Write(b)
}

func (v *Client) readLoop() {
	v.wg.Background(func() {
		for {
			buf := make([]byte, 65535)
			n, _, err := v.conn.ReadFrom(buf)
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					go v.mux.RLock(func() {
						v.call.HandlerUDP(err, nil)
					})
					continue
				}
				return
			}
			if n == 0 {
				continue
			}
			go v.mux.RLock(func() {
				v.call.HandlerUDP(nil, buf[:n])
			})
		}
	})
}

func (v *Client) HandleFunc(h HandlerUDP) {
	v.mux.Lock(func() {
		v.call = h
	})
}
