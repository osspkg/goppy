/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package event

import (
	"encoding/json"
	"fmt"
	"testing"

	"go.osspkg.com/goppy/xtest"
)

func TestUnit_Event(t *testing.T) {
	ev := &Message{}
	err := json.Unmarshal([]byte(`{"e":1001,"u":"1111","d":{"token":"12345","os":"debian"}}`), ev)
	xtest.NoError(t, err)

	b, err := json.Marshal(ev)
	xtest.NoError(t, err)
	xtest.Equal(t, string(b), "{\"e\":1001,\"d\":{\"token\":\"12345\",\"os\":\"debian\"}}")

	ev.Error(fmt.Errorf("error1"))

	b, err = json.Marshal(ev)
	xtest.NoError(t, err)
	xtest.Equal(t, string(b), "{\"e\":1001,\"d\":null,\"err\":\"error1\"}")

	ev.Reset()

	b, err = json.Marshal(ev)
	xtest.NoError(t, err)
	xtest.Equal(t, string(b), "{\"e\":0,\"d\":null}")
}
