package bearerauth_test

import (
	"net/http"
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v3/plugins/auth/bearerauth"
)

func TestUnit_BearerAuth(t *testing.T) {
	h := http.Header{}
	bearerauth.Encode(h, "1")
	val, err := bearerauth.Decode(h)
	casecheck.NoError(t, err)
	casecheck.Equal(t, val, "1")
}
