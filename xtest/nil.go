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
