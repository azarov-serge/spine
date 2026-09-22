package main

import (
	"log"

	"github.com/azarov-serge/spine"
)

func main() {
	app := spine.New(spine.Config{ServiceName: "basic-example"})

	app.Use(
		spine.RequestID(),
		spine.Logger(),
		spine.Recover(),
	)

	app.GET("/health", func(c *spine.Context) error {
		return c.OK(map[string]string{"status": "ok"})
	})

	app.GET("/ping", func(c *spine.Context) error {
		return c.Text(200, "pong")
	})

	app.Group("/api", func(api spine.Router) {
		api.GET("/hello/{name}", func(c *spine.Context) error {
			return c.OK(map[string]string{
				"message": "hello, " + c.Param("name"),
			})
		})
	})

	log.Fatal(app.Run(":8080"))
}
