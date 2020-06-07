package ratelimit

import (
	"strconv"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/patrickmn/go-cache"
)

const pluginName = "ratelimit"

func init() {
	plugin.Register(pluginName, setup)
}

func setup(c *caddy.Controller) error {
	p, err := parseConfig(c)
	if err != nil {
		return plugin.Error(pluginName, err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		p.Next = next
		return p
	})

	c.OnStartup(func() error {
		metrics.MustRegister(c, DropCount)
		return nil
	})

	return nil
}

func parseConfig(c *caddy.Controller) (*RateLimit, error) {
	rl := &RateLimit{
		limit:     defaultRatelimit,
		whitelist: make(map[string]bool),
		buckets:   cache.New(defaultTimeWindow, defaultPurgeInterval),
	}

	for c.Next() {
		args := c.RemainingArgs()

		if len(args) > 0 {
			ratelimit, err := strconv.Atoi(args[0])
			if err != nil {
				return nil, c.ArgErr()
			}
			rl.limit = ratelimit
		}

		for c.NextBlock() {
			switch c.Val() {
			case "whitelist":
				whitelist := c.RemainingArgs()
				for _, ip := range whitelist {
					rl.whitelist[ip] = true
				}

			default:
				return nil, c.ArgErr()
			}
		}
	}
	return rl, nil
}
