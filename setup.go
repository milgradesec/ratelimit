package ratelimit

import (
	"strconv"
	"time"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
)

func init() { plugin.Register("ratelimit", setup) }

func setup(c *caddy.Controller) error {
	p, err := parseRatelimit(c)
	if err != nil {
		return plugin.Error("ratelimit", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		p.Next = next
		return p
	})

	c.OnStartup(func() error {
		m := dnsserver.GetConfig(c).Handler("prometheus")
		if m == nil {
			return nil
		}
		if x, ok := m.(*metrics.Metrics); ok {
			x.MustRegister(ratelimited)
		}
		return nil
	})
	return nil
}

func parseRatelimit(c *caddy.Controller) (*RateLimit, error) {
	r := &RateLimit{
		limit:     defaultRatelimit,
		whitelist: make(map[string]bool),
		bucket:    cache.New(time.Hour, time.Hour),
	}

	for c.Next() {
		args := c.RemainingArgs()

		if len(args) > 0 {
			ratelimit, err := strconv.Atoi(args[0])
			if err != nil {
				return nil, c.ArgErr()
			}
			r.limit = ratelimit
		}

		for c.NextBlock() {
			switch c.Val() {
			case "whitelist":
				whitelist := c.RemainingArgs()
				for _, ip := range whitelist {
					r.whitelist[ip] = true
				}
			}
		}
	}
	return r, nil
}

var (
	ratelimited = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "ratelimit",
		Name:      "dropped_total",
		Help:      "Count of requests that have been dropped because of rate limit.",
	}, []string{"server"})
)
