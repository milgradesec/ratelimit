package ratelimit

import (
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	LimitedCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "ratelimit",
		Name:      "dropped_total",
		Help:      "Count of requests that have been dropped because of rate limit.",
	}, []string{"server"})
)
