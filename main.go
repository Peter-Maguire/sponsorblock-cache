package main

import (
	"crypto/tls"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"github.com/peter-maguire/sponsorblock-cache/entity"
	"io"
	"net/http"
	"time"
)

var client *http.Client

func main() {
	hashCache := cache.New(30*time.Minute, 60*time.Minute)
	e := echo.New()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fmt.Printf("[%s] %s\n", c.Request().Method, c.Request().URL)
			c.Response().Header().Set("access-control-allow-origin", "*")
			c.Response().Header().Set("access-control-allow-methods", "GET, POST, OPTIONS, DELETE")
			c.Response().Header().Set("access-control-allow-headers", "Content-Type, If-None-Match, x-client-name")
			c.Response().Header().Set("access-control-max-age", "86400")
			if c.Request().Method == "GET" {
				c.Response().Header().Set("cache-control", "max-age=300")
			} else if c.Request().Method == "OPTIONS" {
				c.Response().Header().Set("cache-control", "max-age=604800")
			}
			return next(c)
		}
	})

	e.GET("/api/branding/:hash", func(c echo.Context) error {
		hash := c.Param("hash")

		result, exists := hashCache.Get(hash)

		if exists {
			fmt.Println("Cache HIT!")
			return c.JSONBlob(200, result.([]byte))
		}

		req := c.Request()
		b, err := getSponsorBlockResponse(req.Method, req.URL.String(), req.Body, nil)

		if err != nil {
			fmt.Println(err)
			return err
		}

		var videoBranding = make(map[string]entity.VideoBranding)
		err = sonic.Unmarshal(b, &videoBranding)
		if err != nil {
			fmt.Println(err)
			return err
		}

		fixedOutput, err := sonic.Marshal(videoBranding)

		go func() {
			fmt.Println("Caching response for hash", hash)
			hashCache.Set(hash, fixedOutput, 30*time.Minute)
		}()
		return c.JSONBlob(200, fixedOutput)
	})

	e.GET("/api/branding", func(c echo.Context) error {
		video := c.QueryParam("videoID")
		result, exists := hashCache.Get(video)

		if exists {
			fmt.Println("Cache HIT!")
			return c.JSONBlob(200, result.([]byte))
		}

		req := c.Request()
		b, err := getSponsorBlockResponse(req.Method, req.URL.String(), req.Body, nil)

		if err != nil {
			fmt.Println(err)
			return err
		}

		go func() {
			fmt.Println("Caching response for hash", video)
			hashCache.Set(video, b, 30*time.Minute)
		}()
		return c.JSONBlob(200, b)
	})

	e.GET("/api/videoLabels/:hash", func(c echo.Context) error {
		hash := "VL-" + c.Param("hash")

		result, exists := hashCache.Get(hash)

		if exists {
			return c.JSONBlob(200, result.([]byte))
		}

		req := c.Request()
		b, err := getSponsorBlockResponse(req.Method, req.URL.String(), req.Body, nil)

		if err != nil {
			fmt.Println(err)
			return err
		}
		go func() {
			fmt.Println("Caching response for hash", hash)
			hashCache.Set(hash, b, 30*time.Minute)
		}()
		return c.JSONBlob(200, b)
	})

	e.GET("/api/skipSegments/:id", func(c echo.Context) error {
		id := "SS-" + c.Param("id")
		result, exists := hashCache.Get(id)

		if exists {
			fmt.Println("Cache HIT!")
			return c.JSONBlob(200, result.([]byte))
		}

		req := c.Request()
		b, err := getSponsorBlockResponse(req.Method, req.URL.String(), req.Body, nil)

		if err != nil {
			fmt.Println(err)
			return err
		}
		go func() {
			fmt.Println("Caching response for hash", id)
			// TODO: This should cache for longer for older videos, or maybe not cache at all?
			hashCache.Set(id, b, 5*time.Minute)
		}()
		return c.JSONBlob(200, b)
	})

	// No need to forward OPTIONS as we know what's going to happen
	e.OPTIONS("*", func(c echo.Context) error {
		return c.NoContent(204)
	})

	e.Match([]string{"GET", "HEAD", "DELETE", "PUT", "POST", "PATCH"}, "*", func(c echo.Context) error {
		req := c.Request()

		b, err := getSponsorBlockResponse(req.Method, req.URL.String(), req.Body, req.Header)

		if err != nil {
			fmt.Println(err)
			return err
		}

		return c.JSONBlob(200, b)
	})

	e.Logger.Fatal(e.Start(":1323"))
}

func getSponsorBlockResponse(method string, url string, body io.Reader, header http.Header) ([]byte, error) {
	newRequest, err := http.NewRequest(method, fmt.Sprintf("https://sponsor.ajay.app%s", url), body)
	newRequest.Header.Set("User-Agent", "SponsorBlock caching server (https://bi.gp)")
	if body != nil {
		newRequest.Header.Set("Authorization", header.Get("Authorization"))
		newRequest.Header.Set("Cookie", header.Get("Cookie"))
		newRequest.Header.Set("Content-Type", header.Get("Content-Type"))
	}
	fmt.Println(newRequest.URL.String())
	if err != nil {
		return nil, err
	}
	res, err := client.Do(newRequest)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}
