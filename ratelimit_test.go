package ratelimit

import (
	"context"
	"testing"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

func TestSetup(t *testing.T) {
	for i, testcase := range []struct {
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
		c := caddy.NewTestController("dns", testcase.config)
		err := setup(c)
		if err != nil {
			if !testcase.failing {
				t.Fatalf("Test #%d expected no errors, but got: %v", i, err)
			}
			continue
		}
		if testcase.failing {
			t.Fatalf("Test #%d expected to fail but it didn't", i)
		}
	}
}

func Test_ServeDNS(t *testing.T) {
	c := caddy.NewTestController("dns", `ratelimit`)
	plugin, err := parseRatelimit(c)
	if err != nil {
		t.Fatal("Failed to initialize the plugin")
	}
	plugin.Next = test.ErrorHandler()

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	req := new(dns.Msg)
	req.SetQuestion("example.org", dns.TypeA)

	_, err = plugin.ServeDNS(context.TODO(), rec, req)
	if err != nil {
		t.Error(err)
	}
}
func TestRatelimiting(t *testing.T) {
	c := caddy.NewTestController("dns", `ratelimit 1`)
	plugin, err := parseRatelimit(c)
	if err != nil {
		t.Fatal("Failed to initialize the plugin")
	}

	allowed, err := plugin.allowRequest("127.0.0.1")
	if err != nil || !allowed {
		t.Fatal("First request must have been allowed")
	}

	allowed, err = plugin.allowRequest("127.0.0.1")
	if err != nil || allowed {
		t.Fatal("Second request must have been ratelimited")
	}
}

func TestWhitelist(t *testing.T) {
	c := caddy.NewTestController("dns", `ratelimit 1 { 
								whitelist 127.0.0.2 127.0.0.1 127.0.0.125 
								}`)
	plugin, err := parseRatelimit(c)
	if err != nil {
		t.Fatal("Failed to initialize the plugin")
	}

	allowed, err := plugin.allowRequest("127.0.0.1")
	if err != nil || !allowed {
		t.Fatal("First request must have been allowed")
	}

	allowed, err = plugin.allowRequest("127.0.0.1")
	if err != nil || !allowed {
		t.Fatal("Second request must have been allowed due to whitelist")
	}

	allowed, err = plugin.allowRequest("76.42.18.23")
	if err != nil || !allowed {
		t.Fatal("First request must have been allowed")
	}

	allowed, err = plugin.allowRequest("76.42.18.23")
	if err != nil || allowed {
		t.Fatal("Second request must have been blocked")
	}

	for i := 0; i < 10; i++ {
		allowed, err := plugin.allowRequest("192.168.1.171")
		if err != nil || !allowed {
			t.Fatal("First request must have been allowed")
		}
	}
}
