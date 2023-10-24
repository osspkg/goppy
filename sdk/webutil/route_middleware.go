/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package webutil

import (
	"net/http"

	"go.osspkg.com/goppy/sdk/log"
)

// RecoveryMiddleware recovery go panic and write to log
func RecoveryMiddleware(l log.Logger) func(
	func(http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if l != nil {
						l.WithFields(log.Fields{"err": err}).Errorf("Recovered")
					}
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			f(w, r)
		}
	}
}
