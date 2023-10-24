/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jwt_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.osspkg.com/goppy/sdk/auth/jwt"
)

type demoJwtPayload struct {
	ID int `json:"id"`
}

func TestUnit_NewJWT(t *testing.T) {
	conf := make([]jwt.Config, 0)
	conf = append(conf, jwt.Config{ID: "789456", Key: "123456789123456789123456789123456789", Algorithm: jwt.AlgHS256})
	j, err := jwt.New(conf)
	require.NoError(t, err)

	payload1 := demoJwtPayload{ID: 159}
	token, err := j.Sign(&payload1, time.Hour)
	require.NoError(t, err)

	payload2 := demoJwtPayload{}
	head1, err := j.Verify(token, &payload2)
	require.NoError(t, err)

	require.Equal(t, payload1, payload2)

	head2, err := j.Verify(token, &payload2)
	require.NoError(t, err)
	require.Equal(t, head1, head2)
}
