/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package tcp

import (
	"crypto/rand"
	"crypto/tls"
	"net"

	"go.osspkg.com/goppy/xnet"
)

func NewListen(address string, certs ...Cert) (net.Listener, error) {
	address = xnet.CheckHostPort(address)

	if len(certs) == 0 {
		l, err := net.Listen("tcp", address)
		if err != nil {
			return nil, err
		}
		return l, nil
	}

	certificates := make([]tls.Certificate, 0, len(certs))
	for _, c := range certs {
		cert, err := tls.LoadX509KeyPair(c.Cert, c.Key)
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
	return l, nil
}
