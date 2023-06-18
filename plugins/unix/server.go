/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unix

import (
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/osspkg/go-sdk/errors"
	"github.com/osspkg/go-sdk/log"
	"github.com/osspkg/goppy/plugins"
)

type (
	Config struct {
		Path string `yaml:"unix"`
	}
)

func (v *Config) Default() {
	v.Path = "./app.socket"
}

func WithServer() plugins.Plugin {
	return plugins.Plugin{
		Config: &Config{},
		Inject: func(c *Config, l log.Logger) (*srv, Server) {
			s := newServer(c, l)
			return s, s
		},
	}
}

type (
	srv struct {
		config   *Config
		sock     net.Listener
		log      log.Logger
		commands map[string]Handler
		mux      sync.RWMutex
	}

	//Handler unix socket command handler
	Handler func([]byte) ([]byte, error)

	Server interface {
		Command(name string, h Handler)
	}
)

func newServer(c *Config, l log.Logger) *srv {
	return &srv{
		config:   c,
		log:      l,
		commands: make(map[string]Handler),
	}
}

func (v *srv) Up() (err error) {
	if err = os.Remove(v.config.Path); err != nil && !os.IsNotExist(err) {
		err = errors.Wrapf(err, "remove unix socket [unix:%s]", v.config.Path)
		return
	}
	if v.sock, err = net.Listen("unix", v.config.Path); err != nil {
		err = errors.Wrapf(err, "init unix socket [unix:%s]", v.config.Path)
		return
	}

	go v.accept()
	return
}

func (v *srv) Down() error {
	if v.sock != nil {
		return v.sock.Close()
	}
	return nil
}

func (v *srv) Command(name string, h Handler) {
	v.mux.Lock()
	v.commands[name] = h
	v.mux.Unlock()
}

func (v *srv) logError(err error, msg string) {
	if err == nil {
		return
	}

	v.log.WithFields(log.Fields{
		"err": err.Error(),
	}).Errorf(msg)
}

func (v *srv) accept() {
	for {
		fd, err := v.sock.Accept()
		if err != nil {
			v.logError(err, "accept unix socket")
			return
		}
		if err = fd.SetDeadline(time.Now().Add(time.Hour)); err != nil {
			v.logError(err, "unix socket set deadline")
			return
		}
		go v.pump(fd)
	}
}

func (v *srv) pump(rw io.ReadWriteCloser) {
	defer func() {
		if err := rw.Close(); err != nil {
			v.logError(err, "close unix socket request")
		}
	}()

	b, err := readBytes(rw)
	if err != nil {
		v.logError(err, "read unix socket request")
		v.logError(writeError(rw, err), "write unix socket error")
		return
	}

	cmd, data := parse(b)

	v.mux.RLock()
	h, ok := v.commands[cmd]
	v.mux.RUnlock()
	if !ok {
		v.logError(writeError(rw, errInvalidCommand), "write unix socket error")
		return
	}

	out, err := h(data)
	if err != nil {
		v.logError(writeError(rw, err), "write unix socket error")
		return
	}
	v.logError(writeBytes(rw, out), "write unix socket response")
}
