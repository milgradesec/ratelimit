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
	defaultRatelimit = 50
)

type RateLimit struct {
	Next plugin.Handler

	limit     int
	whitelist map[string]bool
	bucket    *cache.Cache
}

func (rl *RateLimit) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	ip := state.IP()

	allow, err := rl.allowRequest(ip)
	if err != nil {
		return dns.RcodeServerFailure, err
	}
	if !allow {
		server := metrics.WithServer(ctx)
		LimitedCount.WithLabelValues(server).Inc()
		return dns.RcodeRefused, nil
	}

	return plugin.NextOrFailure(rl.Name(), rl.Next, ctx, w, r)
}

func (rl *RateLimit) allowRequest(ip string) (bool, error) {
	if ip == "" {
		return false, errors.New("invalid empty ip")
	}

	if strings.HasPrefix(ip, "192.168.1.") {
		return true, nil
	}

	if rl.whitelist[ip] {
		return true, nil
	}

	cached, found := rl.bucket.Get(ip)
	if !found {
		rl.bucket.Set(ip, rate.New(rl.limit, time.Second), time.Hour)
		cached, found = rl.bucket.Get(ip)
		if !found {
			return true, errors.New("cache error: just inserted item disappeared")
		}
	}

	token, check := cached.(*rate.RateLimiter)
	if !check {
		return true, errors.New("cache error: type mismatch")
	}

	allow, _ := token.Try()
	return allow, nil
}

func (rl *RateLimit) Name() string {
	return "ratelimit"
}
