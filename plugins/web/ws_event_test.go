/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnit_Event(t *testing.T) {
	ev := &event{}
	err := json.Unmarshal([]byte(`{"e":1001,"u":"1111","d":{"token":"12345","os":"debian"}}`), ev)
	require.NoError(t, err)

	b, err := json.Marshal(ev)
	require.NoError(t, err)
	require.Equal(t, string(b), "{\"e\":1001,\"d\":{\"token\":\"12345\",\"os\":\"debian\"},\"u\":\"1111\"}")

	ev.Error(fmt.Errorf("error1"))

	b, err = json.Marshal(ev)
	require.NoError(t, err)
	require.Equal(t, string(b), "{\"e\":1001,\"d\":null,\"err\":\"error1\",\"u\":\"1111\"}")

	ev.Reset()

	b, err = json.Marshal(ev)
	require.NoError(t, err)
	require.Equal(t, string(b), "{\"e\":0,\"d\":null}")
}
