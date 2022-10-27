package main

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"custody-merchant-admin/middleware/cache"
)

func main() {
	r := echo.New()

	store := cache.NewInMemoryStore(time.Second)

	// Cached Page
	r.GET("/ping", func(c echo.Context) error {
		c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
		return nil
	})

	r.GET("/cache_ping", cache.CachePage(store, time.Minute, func(c echo.Context) error {
		c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
		return nil
	}))

	// Listen and Server in 0.0.0.0:8080
	r.Start(":8080")
}
