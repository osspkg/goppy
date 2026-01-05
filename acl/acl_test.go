/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl_test

import (
	"context"
	"testing"
	"time"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v3/acl"
)

func TestUnit_NewACL(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := acl.NewInMemoryStorage(map[string][]byte{})

	aclStore := acl.New(ctx, store, acl.Options{
		CacheSize:          10,
		CacheDeadline:      time.Minute,
		CacheCheckInterval: time.Minute,
	})

	email := "demo@example.com"

	t.Log("user not exist")

	levels, err := aclStore.HasMany(email, 1, 2, 3, 4)
	casecheck.Error(t, err)
	casecheck.Nil(t, levels)

	casecheck.Error(t, aclStore.Set(email, 10, 1))

	t.Log("user exist")

	casecheck.NoError(t, store.ChangeACL(email, []byte("")))

	casecheck.NoError(t, aclStore.Set(email, 10, 1))

	levels, err = aclStore.HasMany(email, 1, 2, 3, 4)
	casecheck.NoError(t, err)
	casecheck.Equal(t, map[uint16]bool{1: true, 2: false, 3: false, 4: false}, levels)

	casecheck.NoError(t, aclStore.Set(email, 2, 10))

	levels, err = aclStore.HasMany(email, 1, 2, 3, 4)
	casecheck.NoError(t, err)
	casecheck.Equal(t, map[uint16]bool{1: true, 2: true, 3: false, 4: false}, levels)
}
