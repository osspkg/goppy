package http

import (
	"strings"
	"testing"
)

func TestUnit_NewClient(t *testing.T) {
	cc := &HTTPClientConfig{}
	cc.Default()
	tc := newHTTPClient(cc.Config)

	resp := tc.Create(func(rb HTTPRequestBind) {
		rb.URI("https://raw.githubusercontent.com/dewep-online/fdns-filters/master/domains.txt")
	})
	if resp.Err() != nil {
		t.Fail()
	}
	if resp.Code() != 200 {
		t.Fail()
	}
	if !strings.Contains(string(resp.Body()), "[Adblock Plus 2.0]") {
		t.Fail()
	}

	resp = tc.Create(func(rb HTTPRequestBind) {
		rb.URI("https://fdns.dewep.online/adblock-rules.json")
	})
	if resp.Err() != nil {
		t.Fail()
	}
	if resp.Code() != 200 {
		t.Fail()
	}
	data := make([]string, 0)
	err := resp.JSON(&data)
	if err != nil {
		t.Fail()
	}

	ok := false
	for _, datum := range data {
		if datum == "https://raw.githubusercontent.com/dewep-online/fdns-filters/master/domains.txt" {
			ok = true
		}
	}
	if !ok {
		t.Fail()
	}
}
