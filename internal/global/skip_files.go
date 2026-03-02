package global

import "strings"

func NeedSkipFile(filepath string) bool {
	return strings.HasSuffix(filepath, "_gen.go") ||
		strings.HasSuffix(filepath, "_test.go") ||
		strings.HasSuffix(filepath, "_easyjson.go") ||
		strings.HasSuffix(filepath, "_mock.go") ||
		strings.Contains(filepath, "/vendor/")
}
