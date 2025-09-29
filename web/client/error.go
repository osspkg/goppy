/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package client

import "go.osspkg.com/ioutils/data"

type HTTPError struct {
	Err         error
	Code        int
	ContentType string
	Raw         *data.Buffer
}

func (e *HTTPError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Err.Error()
}
