package search

import "regexp"

var (
	rexIndexName      = regexp.MustCompile(`^[a-z0-9\_]+$`)
	rexUpperCamelCase = regexp.MustCompile(`^[A-Z][A-Za-z0-9]+$`)
)

func isValidIndexName(v string) bool {
	return rexIndexName.MatchString(v)
}

func isUpperCamelCase(v string) bool {
	return rexUpperCamelCase.MatchString(v)
}
