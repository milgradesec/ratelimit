# ratelimit

[![Build Status](https://img.shields.io/travis/milgradesec/ratelimit/master.svg?label=build)](https://travis-ci.org/milgradesec/ratelimit)
[![Go Report Card](https://goreportcard.com/badge/milgradesec/ratelimit)](https://goreportcard.com/badge/github.com/milgradesec/ratelimit)
[![Code Coverage](https://img.shields.io/codecov/c/github/milgradesec/ratelimit/master.svg)](https://codecov.io/github/milgradesec/ratelimit?branch=master)

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
