package ratelimit

import (
	"errors"
	"strings"
	"time"

	"github.com/beefsack/go-rate"
	"github.com/patrickmn/go-cache"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

const (
	defaultRatelimit  = 50
	defaultTimeWindow = 15 * time.Second
)

// RateLimit is a plugin that implements response rate limiting
// using a token bucket algorithm.
type RateLimit struct {
	Next plugin.Handler

	limit     int
	whitelist map[string]bool
	bucket    *cache.Cache
}

// ServeDNS implements the plugin.Handler interface.
func (rl *RateLimit) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	if state.Proto() == "tcp" {
		// No ratelimit is applied for TCP clients,
		// pass the request to the next plugin.
		return plugin.NextOrFailure(rl.Name(), rl.Next, ctx, w, r)
	}

	allow, err := rl.check(state.IP())
	if err != nil {
		return dns.RcodeServerFailure, err
	}

	if !allow {
		DropCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
		return dns.RcodeRefused, nil
	}

	return plugin.NextOrFailure(rl.Name(), rl.Next, ctx, w, r)
}

// Name implements the plugin.Handler interface.
func (rl *RateLimit) Name() string {
	return "ratelimit"
}

func (rl *RateLimit) check(ip string) (bool, error) {
	if strings.HasPrefix(ip, "192.168.1.") {
		return true, nil
	}

	if rl.whitelist[ip] {
		return true, nil
	}

	cached, found := rl.bucket.Get(ip)
	if !found {
		rl.bucket.Set(ip, rate.New(rl.limit, time.Second), cache.DefaultExpiration)
		cached, found = rl.bucket.Get(ip)
		if !found {
			return true, errors.New("cache error: just inserted item disappeared")
		}
	}

	token, ok := cached.(*rate.RateLimiter)
	if !ok {
		return true, errors.New("cache error: type mismatch")
	}

	allow, _ := token.Try()
	return allow, nil
}
