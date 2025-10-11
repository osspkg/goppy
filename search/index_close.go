/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search

import "go.osspkg.com/errors"

func (v *service) Close() error {
	var err error

	for name, index := range v.list.Yield() {
		if e := index.Close(); e != nil {
			err = errors.Wrap(err, errors.Wrapf(e, "close index for '%s'", name))
		}
		v.list.Del(name)
	}

	return err
}
