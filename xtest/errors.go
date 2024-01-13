/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xtest

import "strings"

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

func ErrorContains(t IUnitTest, err error, need string, args ...interface{}) {
	if err == nil {
		t.Helper()
		t.Errorf(errorMessage(args, "Want error, but got <nil>"))
		t.FailNow()
		return
	}

	if strings.Contains(err.Error(), need) {
		return
	}
	t.Helper()
	t.Errorf(errorMessage(args, "Not found\nSearchData: %+v\nNeed: %+v", err.Error(), need))
	t.FailNow()
}
