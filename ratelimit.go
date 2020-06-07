package ratelimit

import (
	"errors"
	"time"

	"github.com/beefsack/go-rate"
	"github.com/patrickmn/go-cache"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

// RateLimit is a plugin that implements response rate limiting
// using a token bucket algorithm.
type RateLimit struct {
	Next plugin.Handler

	limit     int
	whitelist map[string]bool
	buckets   *cache.Cache
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

	if allow {
		return plugin.NextOrFailure(rl.Name(), rl.Next, ctx, w, r)
	}

	DropCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
	return dns.RcodeRefused, nil
}

// Name implements the plugin.Handler interface.
func (rl *RateLimit) Name() string {
	return "ratelimit"
}

// check determines if an ip has surpassed the limit and the request
// has to be refused, otherwise the rate limiter is increased.
func (rl *RateLimit) check(ip string) (bool, error) {
	if rl.whitelist[ip] {
		return true, nil
	}

	item, found := rl.buckets.Get(ip)
	if !found {
		rl.buckets.Set(ip, rate.New(rl.limit, time.Second), cache.DefaultExpiration)
		item, _ = rl.buckets.Get(ip)
	}

	token, ok := item.(*rate.RateLimiter)
	if !ok {
		return true, errors.New("cache error: type mismatch")
	}

	allow, _ := token.Try()
	return allow, nil
}

const (
	defaultRatelimit     = 50
	defaultTimeWindow    = 15 * time.Second
	defaultPurgeInterval = 10 * time.Minute
)
