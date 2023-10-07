/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unixsocket

import (
	"io"
	"net"
	"os"
	"sync"

	"github.com/osspkg/goppy/sdk/errors"
	"github.com/osspkg/goppy/sdk/iosync"
	"github.com/osspkg/goppy/sdk/ioutil"
)

var (
	ErrServAlreadyRunning = errors.New("server already running")
	ErrServAlreadyStopped = errors.New("server already stopped")
)

type (
	CommandHandler func([]byte) ([]byte, error)

	Server struct {
		status   iosync.Switch
		path     string
		socket   net.Listener
		commands map[string]CommandHandler
		mux      sync.RWMutex
		logError func(err error)
	}
)

func NewServer(path string) *Server {
	return &Server{
		path:     path,
		status:   iosync.NewSwitch(),
		commands: make(map[string]CommandHandler),
		logError: func(_ error) {},
	}
}

func (v *Server) ErrorLog(handler func(err error)) {
	v.mux.Lock()
	v.logError = func(err error) {
		if err == nil {
			return
		}
		handler(err)
	}
	v.mux.Unlock()
}

func (v *Server) AddCommand(name string, handler CommandHandler) {
	v.mux.Lock()
	v.commands[name] = handler
	v.mux.Unlock()
}

func (v *Server) Down() error {
	v.mux.Lock()
	defer v.mux.Unlock()

	if !v.status.Off() {
		return ErrServAlreadyStopped
	}
	if v.socket != nil {
		return v.socket.Close()
	}
	return nil
}

func (v *Server) Up() error {
	if !v.status.On() {
		return ErrServAlreadyRunning
	}
	err := os.Remove(v.path)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "remove unix socket")
	}
	if v.socket, err = net.Listen("unix", v.path); err != nil {
		return errors.Wrapf(err, "init unix socket")
	}
	for {
		fd, err := v.socket.Accept()
		if err != nil {
			return err
		}
		go v.handler(fd)
	}
}

func (v *Server) handler(rwc io.ReadWriteCloser) {
	v.mux.RLock()
	defer func() {
		v.mux.RUnlock()
		v.logError(rwc.Close())
	}()

	b, err := ioutil.ReadBytes(rwc, newLine)
	if err != nil {
		v.logError(errors.Wrapf(err, "read unix socket request"))
		v.logError(errors.Wrapf(writeError(rwc, err), "write unix socket error"))
		return
	}
	command, data := parseCommand(b)
	handler, ok := v.commands[command]
	if !ok {
		v.logError(errors.Wrapf(ErrInvalidCommand, command))
		v.logError(errors.Wrapf(writeError(rwc, ErrInvalidCommand), "write unix socket error"))
		return
	}

	out, err := handler(data)
	if err != nil {
		v.logError(errors.Wrapf(err, "call command '%s'", command))
		v.logError(errors.Wrapf(writeError(rwc, err), "write unix socket error"))
		return
	}
	v.logError(errors.Wrapf(ioutil.WriteBytes(rwc, out, newLine), "write unix socket response"))
}