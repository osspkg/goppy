/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package epoll

import (
	"go.osspkg.com/goppy/errors"
)

var (
	errServAlreadyRunning = errors.New("server already running")
	errServAlreadyStopped = errors.New("server already stopped")
	errEpollEmptyEvents   = errors.New("epoll empty event")
)

var (
	defaultEOF = []byte("\r\n")
)
