/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl_test

import (
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v3/acl"
)

func TestUnit_NewInMemoryStorage(t *testing.T) {
	store := acl.NewInMemoryStorage(map[string][]byte{
		"u1": []byte("123"),
		"u2": []byte("456"),
	})
	casecheck.NotNil(t, store)

	val, err := store.FindACL("u1")
	casecheck.NoError(t, err)
	casecheck.Equal(t, []byte{49, 50, 51}, val)

	val, err = store.FindACL("u2")
	casecheck.NoError(t, err)
	casecheck.Equal(t, []byte{52, 53, 54}, val)

	val, err = store.FindACL("u3")
	casecheck.Error(t, err)
	casecheck.Equal(t, []byte{}, val)

	err = store.ChangeACL("u2", []byte("789"))
	casecheck.NoError(t, err)

	val, err = store.FindACL("u2")
	casecheck.NoError(t, err)
	casecheck.Equal(t, []byte{55, 56, 57}, val)

	val, err = store.FindACL("u5")
	casecheck.Error(t, err)
	casecheck.Equal(t, []byte{}, val)

	err = store.ChangeACL("u5", []byte("333"))
	casecheck.NoError(t, err)

	val, err = store.FindACL("u5")
	casecheck.NoError(t, err)
	casecheck.Equal(t, []byte{51, 51, 51}, val)
}
