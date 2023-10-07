/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package epoll

import (
	"net"
	"sync"
	"syscall"

	"github.com/osspkg/goppy/sdk/errors"
	"github.com/osspkg/goppy/sdk/log"
	"github.com/osspkg/goppy/sdk/netutil"
	"golang.org/x/sys/unix"
)

const (
	epollEvents          = unix.POLLIN | unix.POLLRDHUP | unix.POLLERR | unix.POLLHUP | unix.POLLNVAL
	epollEventCount      = 100
	epollEventIntervalMS = 500
)

type (
	epollEventsSlice []unix.EpollEvent

	epoll struct {
		fd     int
		conn   epollNetMap
		events epollEventsSlice
		nets   epollNetSlice
		log    log.Logger
		mux    sync.RWMutex
	}
)

func newEpoll(l log.Logger) (*epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoll{
		fd:     fd,
		conn:   make(epollNetMap),
		events: make(epollEventsSlice, epollEventCount),
		nets:   make(epollNetSlice, epollEventCount),
		log:    l,
	}, nil
}

func (v *epoll) AddOrClose(c net.Conn) error {
	fd := netutil.FileDescriptor(c)
	err := unix.EpollCtl(v.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: epollEvents, Fd: int32(fd)})
	if err != nil {
		return errors.Wrap(err, c.Close())
	}
	v.mux.Lock()
	v.conn[fd] = &epollNetItem{Conn: c, Fd: fd}
	v.mux.Unlock()
	return nil
}

func (v *epoll) removeFD(fd int) error {
	return unix.EpollCtl(v.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (v *epoll) Close(c *epollNetItem) error {
	v.mux.Lock()
	defer v.mux.Unlock()
	return v.closeConn(c)
}

func (v *epoll) closeConn(c *epollNetItem) error {
	if err := v.removeFD(c.Fd); err != nil {
		return err
	}
	delete(v.conn, c.Fd)
	return c.Conn.Close()
}

func (v *epoll) CloseAll() (err error) {
	v.mux.Lock()
	defer v.mux.Unlock()

	for _, conn := range v.conn {
		if err0 := v.closeConn(conn); err0 != nil {
			err = errors.Wrap(err, err0)
		}
	}
	v.conn = make(epollNetMap)
	return
}

func (v *epoll) getConn(fd int) (*epollNetItem, bool) {
	v.mux.RLock()
	conn, ok := v.conn[fd]
	v.mux.RUnlock()
	return conn, ok
}

func (v *epoll) Wait() (epollNetSlice, error) {
	n, err := unix.EpollWait(v.fd, v.events, epollEventIntervalMS)
	if err != nil {
		return nil, err
	}
	if n <= 0 {
		return nil, errEpollEmptyEvents
	}

	v.nets = v.nets[:0]
	for i := 0; i < n; i++ {
		fd := int(v.events[i].Fd)
		conn, ok := v.getConn(fd)
		if !ok {
			if err = v.removeFD(fd); err != nil {
				v.log.WithFields(log.Fields{
					"err": err.Error(),
					"fd":  fd,
				}).Errorf("Close fd")
			}
			continue
		}
		if conn.IsAwait() {
			continue
		}
		conn.Await(true)

		switch v.events[i].Events {
		case unix.POLLIN:
			v.nets = append(v.nets, conn)
		default:
			if err = v.Close(conn); err != nil {
				v.log.WithFields(log.Fields{"err": err.Error()}).Errorf("Epoll close connect")
			}
		}
	}

	return v.nets, nil
}
