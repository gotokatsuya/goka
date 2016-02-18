package main

import (
	"github.com/gotokatsuya/goka"
)

func main() {
	g := goka.New()
	g.SetDebug(true)

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
