package unix

import (
	"bytes"
	"io"

	"github.com/osspkg/go-sdk/errors"
)

var (
	delimstring = "\n"
	delimbyte   = []byte(delimstring)
	delimlen    = len(delimbyte)

	cmddelimstring = " "
	cmddelim       = byte(' ')

	errInvalidCommand = errors.New("command not found")
)

func readBytes(v io.Reader) ([]byte, error) {
	var (
		n   int
		err error
		b   = make([]byte, 0, 512)
	)

	for {
		if len(b) == cap(b) {
			b = append(b, 0)[:len(b)]
		}
		n, err = v.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(b) < delimlen {
			return b, io.EOF
		}
		if bytes.Equal(delimbyte, b[len(b)-delimlen:]) {
			b = b[:len(b)-delimlen]
			break
		}
	}
	return b, nil
}

func writeBytes(v io.Writer, b []byte) error {
	if len(b) < delimlen || !bytes.Equal(delimbyte, b[len(b)-delimlen:]) {
		b = append(b, delimbyte...)
	}
	if _, err := v.Write(b); err != nil {
		return err
	}
	return nil
}

func writeError(v io.Writer, err error) error {
	return writeBytes(v, []byte(err.Error()))
}

func parse(b []byte) (string, []byte) {
	for i := 0; i < len(b); i++ {
		if b[i] == cmddelim {
			if len(b) > i+2 {
				return string(b[0:i]), b[i+1:]
			}
			return string(b[0:i]), nil
		}
	}
	return string(b), nil
}
