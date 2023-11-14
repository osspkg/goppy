package xtest

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

func Equal(t IUnitTest, expected interface{}, actual interface{}, args ...interface{}) {
	et, at := reflect.ValueOf(expected).Kind(), reflect.ValueOf(actual).Kind()
	if et != at {
		t.Helper()
		t.Errorf(errorMessage(args, "Different type\nExpected: %T\nActual: %T", expected, actual))
		t.FailNow()
		return
	}
	ev, av := fmt.Sprintf("%+v", expected), fmt.Sprintf("%+v", actual)
	if ev == av {
		return
	}
	t.Helper()
	t.Errorf(errorMessage(args, "Value is not identical\nExpected: %+v\nActual: %+v", expected, actual))
	t.FailNow()
}

func NotEqual(t IUnitTest, expected interface{}, actual interface{}, args ...interface{}) {
	et, at := reflect.ValueOf(expected).Kind(), reflect.ValueOf(actual).Kind()
	if et != at {
		t.Helper()
		t.Errorf(errorMessage(args, "Different type\nExpected: %T\nActual: %T", expected, actual))
		t.FailNow()
		return
	}
	ev, av := fmt.Sprintf("%+v", expected), fmt.Sprintf("%+v", actual)
	if ev != av {
		return
	}
	t.Helper()
	t.Errorf(errorMessage(args, "Value is not identical\nExpected: %+v\nActual: %+v", expected, actual))
	t.FailNow()
}

func True(t IUnitTest, actual bool, args ...interface{}) {
	if !actual {
		return
	}
	t.Helper()
	t.Errorf(errorMessage(args, "Want <true>, but got: %+v", actual))
	t.FailNow()
}

func False(t IUnitTest, actual bool, args ...interface{}) {
	if actual {
		return
	}
	t.Helper()
	t.Errorf(errorMessage(args, "Want <false>, but got: %+v", actual))
	t.FailNow()
}

func Contains(t IUnitTest, searchData interface{}, need interface{}, args ...interface{}) {
	dt, st := reflect.ValueOf(searchData), reflect.ValueOf(need)
	var (
		found bool
	)

	if s1, s2, ok0 := asString(searchData, need); ok0 {
		found = strings.Contains(s1, s2)
	} else if b1, b2, ok1 := asBytes(searchData, need); ok1 {
		found = bytes.Contains(b1, b2)
	} else if dt.Kind() == reflect.Map {
		for _, value := range dt.MapKeys() {
			if value.Kind() != st.Kind() {
				continue
			}
			if reflect.DeepEqual(value.Interface(), st.Interface()) {
				found = true
				break
			}
		}
	} else {
		t.Helper()
		t.Errorf(errorMessage(args, "Unsupported types\nSearchData: %T\nNeed: %T", searchData, need))
		t.FailNow()
		return
	}

	if found {
		return
	}
	t.Helper()
	t.Errorf(errorMessage(args, "Not found\nSearchData: %+v\nNeed: %+v", searchData, need))
	t.FailNow()
}

func NotContains(t IUnitTest, searchData interface{}, need interface{}, args ...interface{}) {
	dt, st := reflect.ValueOf(searchData), reflect.ValueOf(need)
	var (
		found bool
	)

	if s1, s2, ok0 := asString(searchData, need); ok0 {
		found = strings.Contains(s1, s2)
	} else if b1, b2, ok1 := asBytes(searchData, need); ok1 {
		found = bytes.Contains(b1, b2)
	} else if dt.Kind() == reflect.Map {
		for _, value := range dt.MapKeys() {
			if value.Kind() != st.Kind() {
				continue
			}
			if reflect.DeepEqual(value.Interface(), st.Interface()) {
				found = true
				break
			}
		}
	} else {
		t.Helper()
		t.Errorf(errorMessage(args, "Unsupported types\nSearchData: %T\nNeed: %T", searchData, need))
		t.FailNow()
		return
	}

	if !found {
		return
	}
	t.Helper()
	t.Errorf(errorMessage(args, "Found\nSearchData: %+v\nNeed: %+v", searchData, need))
	t.FailNow()
}

func asString(v0 interface{}, v1 interface{}) (string, string, bool) {
	sv0, ok0 := v0.(string)
	sv1, ok1 := v1.(string)
	return sv0, sv1, ok0 && ok1
}

func asBytes(v0 interface{}, v1 interface{}) ([]byte, []byte, bool) {
	sv0, ok0 := v0.([]byte)
	sv1, ok1 := v1.([]byte)
	return sv0, sv1, ok0 && ok1
}
