# ratelimit

![CI](https://github.com/milgradesec/ratelimit/workflows/CI/badge.svg)
[![CodeQL](https://github.com/milgradesec/ratelimit/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/milgradesec/ratelimit/actions/workflows/codeql-analysis.yml)
[![codecov](https://codecov.io/gh/milgradesec/ratelimit/branch/main/graph/badge.svg)](https://codecov.io/gh/milgradesec/ratelimit)
[![Go Report Card](https://goreportcard.com/badge/milgradesec/ratelimit)](https://goreportcard.com/badge/github.com/milgradesec/ratelimit)
[![Go Reference](https://pkg.go.dev/badge/github.com/milgradesec/ratelimit.svg)](https://pkg.go.dev/github.com/milgradesec/ratelimit)
![GitHub](https://img.shields.io/github/license/milgradesec/filter)

## Description

The _ratelimit_ plugin enables response rate limiting to mitigate DNS attacks.

## Syntax

```corefile
ratelimit LIMIT
```

- **LIMIT** the amount of responses-per-second allowed from an IP.

```corefile
ratelimit LIMIT {
    whitelist [IPs...]
}
```

- `whitelist` the list of IPs exluded from rate limit.

## Metrics

If monitoring is enabled (via the _prometheus_ plugin) then the following metric are exported:

- `coredns_ratelimit_dropped_request_total{server}` - count per server

## Examples

```corefile
ratelimit 50 {
    whitelist 127.0.0.1 192.168.1.25 10.240.1.1
}
```
