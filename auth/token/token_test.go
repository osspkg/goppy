/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package token_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v3/auth/token"
)

type (
	mockModel struct {
		Data mockModelData
	}

	mockModelData struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}
)

func (m *mockModel) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Data)
}

func (m *mockModel) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &m.Data)
}

func TestUnit_JWT(t *testing.T) {
	conf := &token.ConfigGroup{}

	casecheck.Error(t, conf.Validate())
	casecheck.NoError(t, conf.Default())
	casecheck.NoError(t, conf.Validate())

	jwt, err := token.New(conf.JWT)
	casecheck.NoError(t, err)
	casecheck.NotNil(t, jwt)

	req := &mockModel{
		Data: mockModelData{
			Name: "qwerty",
			ID:   123456,
		},
	}

	tokId, tokData, err := jwt.CreateJWT(req, "test:user", time.Hour)
	casecheck.NoError(t, err)
	casecheck.NotNil(t, tokData)
	casecheck.NotEqual(t, tokId, uuid.Nil)

	h, p, err := jwt.VerifyJWT(tokData)
	casecheck.NoError(t, err)
	casecheck.NotNil(t, h)
	casecheck.NotNil(t, p)

	resp := &mockModel{}
	casecheck.NoError(t, resp.UnmarshalJSON(p))

	casecheck.Equal(t, req.Data, resp.Data)
	casecheck.Equal(t, tokId.String(), h.TokenID)
	casecheck.Equal(t, "test:user", h.Audience)

	fmt.Printf("%#v\n", *h)

}
