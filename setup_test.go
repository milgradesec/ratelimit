package ratelimit

import (
	"testing"

	"github.com/caddyserver/caddy"
)

func TestSetup(t *testing.T) {
	for i, test := range []struct {
		config  string
		failing bool
	}{
		{`ratelimit`, false},

		{`ratelimit 100`, false},

		{`ratelimit { 
					whitelist 127.0.0.1
				}`, false},

		{`ratelimit 50 {
					whitelist 127.0.0.1 176.103.130.130
				}`, false},

		{`ratelimit error`, true},
	} {
		c := caddy.NewTestController("dns", test.config)
		err := setup(c)
		if err != nil {
			if !test.failing {
				t.Fatalf("Test #%d expected no errors, but got: %v", i, err)
			}
			continue
		}
		if test.failing {
			t.Fatalf("Test #%d expected to fail but it didn't", i)
		}
	}
}
