package internal

import (
	"io"
)

func ReadAll(r io.ReadCloser) ([]byte, error) {
	defer r.Close() //nolint: errcheck
	return io.ReadAll(r)
}
