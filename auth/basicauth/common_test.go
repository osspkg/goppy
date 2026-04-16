package basicauth_test

import (
	"net/http"
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v3/auth/basicauth"
)

func TestUnit_BasicAuth(t *testing.T) {
	h := http.Header{}
	basicauth.Encode(h, "1", "2")
	l, p, err := basicauth.Decode(h)
	casecheck.NoError(t, err)
	casecheck.Equal(t, l, "1")
	casecheck.Equal(t, p, "2")
}
