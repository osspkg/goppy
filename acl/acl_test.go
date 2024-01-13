/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl_test

import (
	"testing"

	acl2 "go.osspkg.com/goppy/acl"
	"go.osspkg.com/goppy/xtest"
)

func TestUnit_NewACL(t *testing.T) {
	store := acl2.NewInMemoryStorage()
	acl := acl2.NewACL(store, 3)

	email := "demo@example.com"

	t.Log("user not exist")

	levels, err := acl.GetAll(email)
	xtest.Error(t, err)
	xtest.Nil(t, levels)

	xtest.Error(t, acl.Set(email, 10, 1))

	t.Log("user exist")

	xtest.NoError(t, store.ChangeACL(email, ""))

	xtest.Error(t, acl.Set(email, 10, 1))

	levels, err = acl.GetAll(email)
	xtest.NoError(t, err)
	xtest.Equal(t, []uint8{0, 0, 0}, levels)

	xtest.NoError(t, acl.Set(email, 2, 10))

	levels, err = acl.GetAll(email)
	xtest.NoError(t, err)
	xtest.Equal(t, []uint8{0, 0, 9}, levels)
}
