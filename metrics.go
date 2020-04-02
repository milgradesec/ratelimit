package ratelimit

import (
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RateLimitCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: pluginName,
		Name:      "dropped_request_total",
		Help:      "Count of requests that have been dropped because of rate limit.",
	}, []string{"server"})
)
