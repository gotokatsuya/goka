# goka

Use fasthttp. Based on echo.

```go
package main

import (
	"github.com/gotokatsuya/goka"
)

func main() {
	g := goka.New()

	g.Get("/hi", func(c *goka.Context) error {
		type User struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		response := make(map[string]interface{})
		response["user"] = &User{ID: 1, Name: "goka"}
		return c.JSON(200, response)
	})

	g.Run(":8080")
}
```

- Response
```
{"user":{"id":1,"name":"goka"}}
```

## Benchmark

Using [wrk](https://github.com/wg/wrk).

### goka

```
$ go run bench/bench.go -waf=goka
$ wrk -t4 -c100 -d30S --timeout 2000 "http://127.0.0.1:8080/welcome"
```

```
Running 30s test @ http://127.0.0.1:8080/welcome
  4 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.16ms  244.33us  12.25ms   87.17%
    Req/Sec    21.68k     1.81k   27.57k    67.61%
  2597468 requests in 30.10s, 656.44MB read
Requests/sec:  86288.27
Transfer/sec:     21.81MB
```

### gin

```
$ go run bench/bench.go -waf=gin
$ wrk -t4 -c100 -d30S --timeout 2000 "http://127.0.0.1:8080/welcome"
```

```
Running 30s test @ http://127.0.0.1:8080/welcome
  4 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   109.05ms  408.31ms   2.69s    93.50%
    Req/Sec    13.67k     2.81k   26.11k    79.09%
  1449432 requests in 30.05s, 342.81MB read
Requests/sec:  48230.39
Transfer/sec:     11.41MB
```
