package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"io"
	"net/http"
	"time"
)

func main() {
	hashCache := cache.New(30*time.Minute, 60*time.Minute)
	e := echo.New()

	e.GET("/api/branding/:hash", func(c echo.Context) error {
		fmt.Printf("[%s] %s\n", c.Request().Method, c.Request().URL)

		hash := c.Param("hash")

		result, exists := hashCache.Get(hash)

		if exists {
			fmt.Println("Cache HIT!")
			return c.JSONBlob(200, result.([]byte))
		}

		req := c.Request()
		newRequest, err := http.NewRequest(req.Method, fmt.Sprintf("https://sponsor.ajay.app%s", c.Request().URL), c.Request().Body)
		newRequest.Header.Set("User-Agent", "Sponsorblock caching server (https://bi.gp)")
		fmt.Println(newRequest.URL.String())
		if err != nil {
			fmt.Println(err)
			return err
		}
		res, err := http.DefaultClient.Do(newRequest)
		if err != nil {
			fmt.Println(err)
			return err
		}

		b, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return err
		}

		c.Response().Header().Set("access-control-allow-origin", "*")
		c.Response().Header().Set("access-control-allow-methods", "GET, POST, DELETE")
		c.Response().Header().Set("access-control-allow-headers", "Content-Type")
		c.Response().Header().Set("access-control-max-age", "86400")
		fmt.Println("Caching response for hash", hash)
		hashCache.Set(hash, b, 30*time.Minute)
		return c.JSONBlob(200, b)
	})

	e.Any("*", func(c echo.Context) error {
		fmt.Printf("[%s] %s\n", c.Request().Method, c.Request().URL)
		req := c.Request()

		newRequest, err := http.NewRequest(req.Method, fmt.Sprintf("https://sponsor.ajay.app%s", c.Request().URL), c.Request().Body)
		newRequest.Header.Set("User-Agent", "Sponsorblock caching server (https://bi.gp)")
		fmt.Println(newRequest.URL.String())
		if err != nil {
			fmt.Println(err)
			return err
		}
		res, err := http.DefaultClient.Do(newRequest)
		if err != nil {
			fmt.Println(err)
			return err
		}

		b, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Println("Response", string(b))

		c.Response().Header().Set("access-control-allow-origin", "*")
		c.Response().Header().Set("access-control-allow-methods", "GET, POST, DELETE")
		c.Response().Header().Set("access-control-allow-headers", "Content-Type")
		c.Response().Header().Set("access-control-max-age", "86400")

		return c.JSONBlob(200, b)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
