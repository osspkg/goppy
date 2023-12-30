/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package server

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
)

type (
	Listen struct {
		conn net.Listener
		tls  bool
	}
)

func NewListen(port int, certs ...Cert) (*Listen, error) {
	address := fmt.Sprintf("0.0.0.0:%d", port)

	if len(certs) == 0 {
		l, err := net.Listen("tcp", address)
		if err != nil {
			return nil, err
		}
		return &Listen{conn: l}, nil
	}

	certificates := make([]tls.Certificate, 0, len(certs))
	for _, c := range certs {
		cert, err := tls.LoadX509KeyPair(c.Public, c.Private)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, cert)
	}
	config := tls.Config{Certificates: certificates, Rand: rand.Reader}
	l, err := tls.Listen("tcp", address, &config)
	if err != nil {
		return nil, err
	}
	return &Listen{conn: l, tls: true}, nil
}

func (v *Listen) Close() error {
	return v.conn.Close()
}

func (v *Listen) Accept() (net.Conn, error) {
	return v.conn.Accept()
}
