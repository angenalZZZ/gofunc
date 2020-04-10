package limiter

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/http/fast"
	"net/http"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	// Request times
	times, n := 10000, 0
	rq := f.TimesRepeat(times, http.StatusOK)
	f.TimesRepeatAppend(rq, 100, http.StatusTooManyRequests)

	app := fast.New()
	rl := NewRateLimiterPerSecond(times)
	app.Get("/", rl.Wrap(func(c *fast.Ctx) {
		c.SendString(c.C.URI().String())
	}))

	// start test
	for _, want := range rq {
		n++
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://google.com?x=%d", n), nil)
		req.Header.Set("x", f.ToString(n))
		res, _ := app.Test(req)
		if want != res.StatusCode {
			t.Error(ErrTooManyRequests)
		}
	}

	// wait the interval, should be good for more
	time.Sleep(time.Second)
	for _, want := range rq {
		n++
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://google.com?x=%d", n), nil)
		req.Header.Set("x", f.ToString(n))
		res, _ := app.Test(req)
		if want != res.StatusCode {
			t.Error(ErrTooManyRequests)
		}
	}

	// many concurrent requests
	rq = f.TimesRepeat(2000, http.StatusOK)
	time.Sleep(time.Second)
	for _, want := range rq {
		n++
		go func(n int, want int) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://google.com?x=%d", n), nil)
			req.Header.Set("x", f.ToString(n))
			res, _ := app.Test(req)
			if res == nil || want != res.StatusCode {
				t.Error(ErrTooManyRequests)
			}
		}(n, want.(int))
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	// Request times
	times, n := 100000, 0
	rq := f.TimesRepeat(times, http.StatusOK)
	f.TimesRepeatAppend(rq, 100, http.StatusTooManyRequests)

	app := fast.New()
	app.Use(New(Config{
		Max: int64(times),
		Handler: func(c *fast.Ctx) {
			c.Status(http.StatusTooManyRequests).SendString(c.C.URI().String())
		},
	}))
	app.Get("/", func(c *fast.Ctx) {
		c.SendString(c.C.URI().String())
	})

	// start test
	for _, want := range rq {
		n++
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://google.com?x=%d", n), nil)
		req.Header.Set("x", f.ToString(n))
		res, _ := app.Test(req)
		if want != res.StatusCode {
			t.Error(ErrTooManyRequests)
		}
	}

	// wait the interval, should be good for more
	time.Sleep(1 * time.Second)
	for _, want := range rq {
		n++
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://google.com?x=%d", n), nil)
		req.Header.Set("x", f.ToString(n))
		res, _ := app.Test(req)
		if want != res.StatusCode {
			t.Error(ErrTooManyRequests)
		}
	}

	// many concurrent requests
	rq = f.TimesRepeat(10000, http.StatusOK)
	time.Sleep(time.Second)
	for _, want := range rq {
		n++
		go func(n int, want int) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://google.com?x=%d", n), nil)
			req.Header.Set("x", f.ToString(n))
			res, _ := app.Test(req)
			if res == nil || want != res.StatusCode {
				t.Error(ErrTooManyRequests)
			}
		}(n, want.(int))
	}
}
