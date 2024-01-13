/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package client

type nilHandler struct{}

func newNilHandler() HandlerUDP {
	return &nilHandler{}
}

func (v *nilHandler) HandlerUDP(_ error, _ []byte) {}
