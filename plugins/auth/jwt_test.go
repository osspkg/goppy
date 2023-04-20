package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type demoJwtPayload struct {
	ID int `json:"id"`
}

func TestUnit_newJWT(t *testing.T) {
	conf := &ConfigJWT{}
	err := conf.Validate()
	require.Error(t, err)

	conf.Default()

	j, err := newJWT(conf.JWT)
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
