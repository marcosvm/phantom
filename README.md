# Phantom

Phantom listens on a port and exposes prometheus metrics about requests being posted to it.

The response will always be 200, the body content will be read and ignored.

The service will record into a counter the contents of the `X-Forwarded-For` header as labels called `origin` and `proxies` if available.

The address to listen to, the header to be used as origin and log levels are configurable, see usage below.

## Example
```
# HELP metrics_posts_received_total The total number of received posts for metrics
# TYPE metrics_posts_received_total counter
# HELP metrics_posts_received_total The total number of received posts for metrics
# TYPE metrics_posts_received_total counter
metrics_posts_received_total{origin="1.2.3.4",proxies=""} 1
metrics_posts_received_total{origin="1.2.3.4",proxies="5.6.7.8"} 1
metrics_posts_received_total{origin="unknown",proxies=""} 1
```

## Usage

```bash
Usage of phantom:
  -header string
    	request address header (default "X-Forwarded-For")
  -listen string
        ip:port for listening to web requests (default ":7777")
  -log.level string
    	debug, info, warn, error (default "info")
```

## Local Build

```bash
go mod download && go build
```

## Docker local build
```Bash
docker build . -t phantom:local
```

## Docker local run
```bash
docker run --rm -p 7777:7777 phanton:local
```

## Testing
```bash
go test -v
=== RUN   TestHandler
level=debug msg="request received from origin" origin=10.1.1.1
--- PASS: TestHandler (0.00s)
=== RUN   TestEmptyBody
--- PASS: TestEmptyBody (0.00s)
PASS
ok  	github.com/marcosvm/phantom	0.316s
```

and
```
go test -bench -cpu=1,2,4,8,16,32 -run=^$ -bench ^BenchmarkPost$ github.com/marcosvm/phantom
goos: darwin
goarch: arm64
pkg: github.com/marcosvm/phantom
BenchmarkPost-10    	 2593036	       464.3 ns/op	    1135 B/op	       7 allocs/op
PASS
ok  	github.com/marcosvm/phantom	1.830s
```

## Production and other environments
Please follow the best practices and workflows for building and publishing.
