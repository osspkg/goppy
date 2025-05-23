/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl_test

import (
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v2/acl"
)

func TestUnit_NewInConfigStorage(t *testing.T) {
	conf := &acl.ConfigInConfigStorage{ACL: map[string]string{
		"u1": "123",
		"u2": "456",
	}}
	store := acl.NewInConfigStorage(conf)
	casecheck.NotNil(t, store)

	val, err := store.FindACL("u1")
	casecheck.NoError(t, err)
	casecheck.Equal(t, "123", val)

	val, err = store.FindACL("u2")
	casecheck.NoError(t, err)
	casecheck.Equal(t, "456", val)

	val, err = store.FindACL("u3")
	casecheck.Error(t, err)
	casecheck.Equal(t, "", val)

	err = store.ChangeACL("u2", "789")
	casecheck.Error(t, err)

	err = store.ChangeACL("u5", "333")
	casecheck.Error(t, err)
}
