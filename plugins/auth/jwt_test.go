package auth_test

import (
	"testing"

	"github.com/deweppro/goppy/plugins/auth"
	"github.com/stretchr/testify/require"
)

func TestUnit_ConfigJWT(t *testing.T) {
	conf := &auth.ConfigJWT{}

	err := conf.Validate()
	require.Error(t, err)

	conf.Default()

	err = conf.Validate()
	require.NoError(t, err)
}
