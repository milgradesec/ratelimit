package ratelimit

import (
	"context"
	"testing"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

func Test_ServeDNS(t *testing.T) {
	c := caddy.NewTestController("dns", `ratelimit 1 {
		whitelist 127.0.0.1
		}`)

	rl, err := parseRatelimit(c)
	if err != nil {
		t.Fatal(err)
	}
	rl.Next = test.NextHandler(dns.RcodeSuccess, nil)

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	m := new(dns.Msg)
	m.SetQuestion("example.org", dns.TypeA)

	rcode, err := rl.ServeDNS(context.TODO(), rec, m)
	if err != nil {
		t.Fatal(err)
	}
	if rcode != dns.RcodeSuccess {
		t.Fatal("First request must have been allowed")
	}

	rcode, err = rl.ServeDNS(context.TODO(), rec, m)
	if err != nil {
		t.Fatal(err)
	}
	if rcode != dns.RcodeRefused {
		t.Fatal("Second request must have been refused")
	}
}

func TestWhitelist(t *testing.T) {
	c := caddy.NewTestController("dns", `ratelimit 1 { 
		whitelist 127.0.0.2 127.0.0.1 127.0.0.125 
		}`)
	rl, err := parseRatelimit(c)
	if err != nil {
		t.Fatal("Failed to initialize the plugin")
	}

	allowed, err := rl.check("127.0.0.1")
	if err != nil || !allowed {
		t.Fatal("First request must have been allowed")
	}

	allowed, err = rl.check("127.0.0.1")
	if err != nil || !allowed {
		t.Fatal("Second request must have been allowed due to whitelist")
	}

	allowed, err = rl.check("76.42.18.23")
	if err != nil || !allowed {
		t.Fatal("First request must have been allowed")
	}

	allowed, err = rl.check("76.42.18.23")
	if err != nil || allowed {
		t.Fatal("Second request must have been blocked")
	}
}
