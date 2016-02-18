package main

import (
	"flag"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gotokatsuya/goka"
)

func main() {
	waf := flag.String("waf", "goka", "Web Application Framework")
	flag.Parse()

	switch *waf {
	case "gin":
		runGin()
	default:
		runGoka()
	}
}

func runGin() {
	r := gin.Default()
	r.GET("/welcome", func(c *gin.Context) {
		c.JSON(http.StatusOK, `[{"name": "Chris McCord"}, {"name": "Matt Sears"}, {"name": "David Stump"}, {"name": "Ricardo Thompson"}]`)
	})
	r.Run(":8080") // listen and serve on 0.0.0.0:8080
}

func runGoka() {
	g := goka.New()
	g.Get("/welcome", func(c *goka.Context) error {
		return c.JSON(http.StatusOK, `[{"name": "Chris McCord"}, {"name": "Matt Sears"}, {"name": "David Stump"}, {"name": "Ricardo Thompson"}]`)
	})
	g.Run(":8080") // listen and serve on 0.0.0.0:8080
}
