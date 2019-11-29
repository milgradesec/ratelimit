package ratelimit

import (
	"errors"
	"strings"
	"time"

	"github.com/beefsack/go-rate"
	"github.com/patrickmn/go-cache"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

type RateLimit struct {
	Next plugin.Handler

	limit     int
	whitelist map[string]bool
	bucket    *cache.Cache
}

const (
	defaultRatelimit    = 50
	defaultResponseSize = 1024
)

func (rl *RateLimit) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	ip := state.IP()

	allow, err := rl.allowRequest(ip)
	if err != nil {
		return dns.RcodeServerFailure, err
	}
	if !allow {
		server := metrics.WithServer(ctx)
		ratelimited.WithLabelValues(server).Inc()
		return dns.RcodeRefused, nil
	}

	rw := dnstest.NewRecorder(w)
	rcode, err := plugin.NextOrFailure(rl.Name(), rl.Next, ctx, rw, r)

	size := rw.Len
	if size > defaultResponseSize && state.Proto() == "udp" {
		for i := 0; i < size/defaultResponseSize; i++ {
			_, err = rl.allowRequest(ip)
			if err != nil {
				return dns.RcodeServerFailure, err
			}
		}
	}
	return rcode, err
}

func (rl *RateLimit) allowRequest(ip string) (bool, error) {
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
