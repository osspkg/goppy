/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package iofile

import (
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/osspkg/goppy/sdk/errors"
)

func IsValidHash(filename string, h hash.Hash, valid string) error {
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	if _, err = io.Copy(h, r); err != nil {
		return errors.Wrapf(err, "calculate file hash")
	}
	result := hex.EncodeToString(h.Sum(nil))
	h.Reset()
	if result != valid {
		return fmt.Errorf("invalid hash: expected[%s] actual[%s]", valid, result)
	}
	return nil
}

func Hash(filename string, h hash.Hash) (string, error) {
	r, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(h, r); err != nil {
		return "", errors.Wrapf(err, "calculate file hash")
	}
	result := hex.EncodeToString(h.Sum(nil))
	h.Reset()
	return result, nil
}
