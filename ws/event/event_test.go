/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package event

import (
	"encoding/json"
	"fmt"
	"testing"

	"go.osspkg.com/casecheck"
)

func TestUnit_Event(t *testing.T) {
	type data struct {
		ID int `json:"id"`
	}

	ev := poolEvents.Get()
	ev.WithID(1)
	err := ev.Encode(&data{ID: 123})
	casecheck.NoError(t, err)

	b, err := json.Marshal(ev)
	casecheck.NoError(t, err)
	casecheck.NotNil(t, b)
	casecheck.Equal(t, `{"e":1,"d":{"id":123}}`, string(b))

	ev.Reset()
	err = json.Unmarshal([]byte("{}"), &ev)
	casecheck.NoError(t, err)
	casecheck.Equal(t, Id(0), ev.Id)
	casecheck.Equal(t, json.RawMessage(nil), ev.Data)
	casecheck.Equal(t, (*string)(nil), ev.Err)

	ev.Reset()
	err = json.Unmarshal([]byte(`{"e":1,"d":{"id":123}}`), &ev)
	casecheck.NoError(t, err)
	casecheck.Equal(t, Id(1), ev.ID())
	casecheck.Equal(t, Id(1), ev.Id)
	casecheck.Equal(t, `{"id":123}`, string(ev.Data))
	casecheck.Equal(t, (*string)(nil), ev.Err)

	d := data{}
	err = ev.Decode(&d)
	casecheck.NoError(t, err)
	casecheck.Equal(t, 123, d.ID)

	d = data{}
	ev.WithError(fmt.Errorf("1"))
	err = ev.Decode(&d)
	casecheck.Error(t, err)
	casecheck.Equal(t, "1", err.Error())
	casecheck.Equal(t, 0, d.ID)

	ev.Reset()
	err = json.Unmarshal([]byte(`{"e":1,"d":{"id":123}}`), &ev)
	casecheck.NoError(t, err)

	d = data{}
	ev.WithError(nil)
	err = ev.Decode(&d)
	casecheck.NoError(t, err)
	casecheck.Equal(t, 123, d.ID)

	err = ev.Encode(func() {})
	casecheck.Error(t, err)
	casecheck.Equal(t, "json: unsupported type: func()", err.Error())
}
