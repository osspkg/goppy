/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl_test

import (
	"testing"

	"go.osspkg.com/goppy/acl"
	"go.osspkg.com/goppy/xtest"
)

func TestUnit_NewInConfigStorage(t *testing.T) {
	conf := &acl.ConfigInConfigStorage{ACL: map[string]string{
		"u1": "123",
		"u2": "456",
	}}
	store := acl.NewInConfigStorage(conf)
	xtest.NotNil(t, store)

	val, err := store.FindACL("u1")
	xtest.NoError(t, err)
	xtest.Equal(t, "123", val)

	val, err = store.FindACL("u2")
	xtest.NoError(t, err)
	xtest.Equal(t, "456", val)

	val, err = store.FindACL("u3")
	xtest.Error(t, err)
	xtest.Equal(t, "", val)

	err = store.ChangeACL("u2", "789")
	xtest.Error(t, err)

	err = store.ChangeACL("u5", "333")
	xtest.Error(t, err)
}
