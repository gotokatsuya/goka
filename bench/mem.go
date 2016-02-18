package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/gotokatsuya/goka"
)

var (
	ginRouter  *gin.Engine
	gokaRouter *goka.Goka
)

func main() {
	waf := flag.String("waf", "goka", "Web Application Framework")
	num := flag.Int("n", 5, "Route Number")
	flag.Parse()

	fmt.Println("Routes:", num)

	memStats := new(runtime.MemStats)
	runtime.GC()
	runtime.ReadMemStats(memStats)
	before := memStats.HeapAlloc

	switch *waf {
	case "gin":
		ginRouter = loadGin(*num)
	default:
		gokaRouter = loadGoka(*num)
	}

	runtime.GC()
	runtime.ReadMemStats(memStats)
	after := memStats.HeapAlloc
	fmt.Println(*waf+":", after-before, "Bytes")
}

func loadGin(num int) *gin.Engine {
	r := gin.Default()
	for i := 0; i < num; i++ {
		r.GET(fmt.Sprintf("/welcome/%d", i), func(c *gin.Context) {
			c.JSON(http.StatusOK, `[{"name": "Chris McCord"}, {"name": "Matt Sears"}, {"name": "David Stump"}, {"name": "Ricardo Thompson"}]`)
		})
	}
	return r
}

func loadGoka(num int) *goka.Goka {
	g := goka.New()
	for i := 0; i < num; i++ {
		g.Get(fmt.Sprintf("/welcome/%d", i), func(c *goka.Context) error {
			return c.JSON(http.StatusOK, `[{"name": "Chris McCord"}, {"name": "Matt Sears"}, {"name": "David Stump"}, {"name": "Ricardo Thompson"}]`)
		})
	}
	return g
}
