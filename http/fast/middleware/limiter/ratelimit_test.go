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
	times, n := 1000, 0
	rl := NewRateLimiterPerSecond(times)
	rq := make([]int, times)
	for i, _ := range rq {
		rq[i] = http.StatusOK
	}
	// StatusCode 429: StatusTooManyRequests
	rq = append(rq, 429, 429, 429, 429, 429, 429, 429, 429, 429, 429)
	app := fast.New()
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
}
