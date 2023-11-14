/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xtest

func NoError(t IUnitTest, err error, args ...interface{}) {
	if err == nil {
		return
	}
	t.Helper()
	t.Errorf(errorMessage(args, "Want <nil>, but got error: %+v", err.Error()))
	t.FailNow()
}

func Error(t IUnitTest, err error, args ...interface{}) {
	if err != nil {
		return
	}
	t.Helper()
	t.Errorf(errorMessage(args, "Want error, but got <nil>"))
	t.FailNow()
}
