/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/goppy/v2/web"
)

type (
	TestModel struct {
		Value struct {
			Name string `json:"name"`
		}
	}
)

func (v *TestModel) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &v.Value)
}

func TestUnit_NewClientHttp_JSON(t *testing.T) {
	model := TestModel{}
	cli := web.NewClientHttp()
	err := cli.Call(context.TODO(), http.MethodGet, "https://osspkg.com/manifest.json", nil, &model)
	casecheck.NoError(t, err)
	casecheck.Equal(t, "OSSPkg", model.Value.Name)
}

func TestUnit_NewClientHttp_Bytes(t *testing.T) {
	var model []byte
	cli := web.NewClientHttp()
	err := cli.Call(context.TODO(), http.MethodGet, "https://osspkg.com/manifest.json", nil, &model)
	casecheck.NoError(t, err)
	casecheck.Contains(t, string(model), "\"OSSPkg\"")
}
