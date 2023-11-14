package xtest

import "fmt"

type IUnitTest interface {
	Errorf(format string, args ...interface{})
	Helper()
	FailNow()
}

func errorMessage(message []interface{}, errMsg string, args ...interface{}) string {
	var msg string
	switch len(message) {
	case 0:
		break
	case 1:
		msg = fmt.Sprintf("%+v", message[0])
	default:
		msg = fmt.Sprintf(fmt.Sprintf("%+v", message[0]), message[1:]...)
	}

	out := fmt.Sprintf("\n[Error] "+errMsg, args...)
	if len(msg) > 0 {
		out += "\n[Message] " + msg
	}
	return out
}
