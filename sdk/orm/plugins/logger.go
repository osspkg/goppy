/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package plugins

import (
	"github.com/osspkg/goppy/sdk/log"
)

var (
	//StdOutLog simple stdout debug log
	StdOutLog = func() log.Logger {
		l := log.Default()
		l.SetLevel(log.LevelDebug)
		l.SetOutput(StdOutWriter)
		return l
	}()
)
