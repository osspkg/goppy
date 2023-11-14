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
