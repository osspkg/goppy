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

func TestUnit_NewACL(t *testing.T) {
	store := acl.NewInMemoryStorage()
	aclStore := acl.New(store, 3)

	email := "demo@example.com"

	t.Log("user not exist")

	levels, err := aclStore.GetAll(email)
	casecheck.Error(t, err)
	casecheck.Nil(t, levels)

	casecheck.Error(t, aclStore.Set(email, 10, 1))

	t.Log("user exist")

	casecheck.NoError(t, store.ChangeACL(email, ""))

	casecheck.Error(t, aclStore.Set(email, 10, 1))

	levels, err = aclStore.GetAll(email)
	casecheck.NoError(t, err)
	casecheck.Equal(t, []uint8{0, 0, 0}, levels)

	casecheck.NoError(t, aclStore.Set(email, 2, 10))

	levels, err = aclStore.GetAll(email)
	casecheck.NoError(t, err)
	casecheck.Equal(t, []uint8{0, 0, 9}, levels)
}
