/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"context"
	"testing"

	"go.osspkg.com/casecheck"
)

type mockModel struct {
	D []byte
}

func (v *mockModel) MarshalJSON() ([]byte, error) {
	return v.D, nil
}

func (v *mockModel) UnmarshalJSON(arg []byte) error {
	v.D = arg
	return nil
}

func TestUnit_Caller(t *testing.T) {
	rpc := NewCaller()

	casecheck.NoError(t, rpc.Add("base.user", func(_ context.Context, m *mockModel) (*mockModel, error) {
		return m, nil
	}))
}
