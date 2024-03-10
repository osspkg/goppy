/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package sqlcommon

import (
	"go.osspkg.com/goppy/xlog"
)

var (
	// StdOutLog simple stdout debug log
	StdOutLog = func() xlog.Logger {
		l := xlog.Default()
		l.SetLevel(xlog.LevelDebug)
		l.SetOutput(StdOutWriter)
		return l
	}()
)
