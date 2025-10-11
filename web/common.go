/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"fmt"
	"strings"
	"time"

	"go.osspkg.com/errors"
)

const (
	defaultTimeout         = 10 * time.Second
	defaultShutdownTimeout = 1 * time.Second
	defaultNetwork         = "tcp"
)

var (
	errServAlreadyRunning = errors.New("server already running")
	errServAlreadyStopped = errors.New("server already stopped")
	errFailContextKey     = errors.New("context key is not found")
)

var (
	networkType = map[string]struct{}{
		"tcp":        {},
		"tcp4":       {},
		"tcp6":       {},
		"unixpacket": {},
		"unix":       {},
	}
)

/**********************************************************************************************************************/

const urlSplitSeparate = "/"

func urlSplit(uri string) []string {
	vv := strings.Split(strings.ToLower(uri), urlSplitSeparate)
	for i := 0; i < len(vv); i++ {
		if len(vv[i]) == 0 {
			copy(vv[i:], vv[i+1:])
			vv = vv[:len(vv)-1]
			i--
		}
	}
	return vv
}

/**********************************************************************************************************************/

func typingJSONKey(v any) string {
	switch vv := v.(type) {
	case string:
		return vv
	case []byte:
		return string(vv)
	case fmt.Stringer:
		return vv.String()
	case fmt.GoStringer:
		return vv.GoString()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func typingJSONValue(v any) any {
	switch vv := v.(type) {
	case string:
		return vv
	case []byte:
		return string(vv)
	case fmt.Stringer:
		return vv.String()
	case fmt.GoStringer:
		return vv.GoString()
	default:
		return v
	}
}
