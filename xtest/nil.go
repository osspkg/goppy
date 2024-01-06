/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xtest

import (
	"reflect"
)

func Nil(t IUnitTest, actual interface{}, args ...interface{}) {
	if isNil(actual) {
		return
	}
	t.Helper()
	t.Errorf(errorMessage(args, "Want <nil>, but got %+v", actual))
	t.FailNow()
}

func NotNil(t IUnitTest, actual interface{}, args ...interface{}) {
	if !isNil(actual) {
		return
	}
	t.Helper()
	t.Errorf(errorMessage(args, "Want not <nil>, but got %+v", actual))
	t.FailNow()
}

func isNil(value interface{}) bool {
	if value == nil {
		return true
	}
	return reflect.ValueOf(value).IsNil()
}
