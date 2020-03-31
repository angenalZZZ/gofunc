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
	app := fast.New()
	times := 10000 // request times
	rl := NewRateLimiter(int64(times), "too many requests")
	rq := make([]int, times)
	for i, _ := range rq {
		rq[i] = http.StatusOK
	}
	rq = append(rq, rl.StatusCode) // StatusCode: StatusTooManyRequests
	allow, deny := func(c *fast.Ctx) {
		c.SendString(c.C.URI().String())
	}, func(c *fast.Ctx) {
		c.Status(rl.StatusCode).SendString(rl.Message + " ?x=" + c.GetHeader("x"))
	}
	app.Get("/", rl.Wrap(allow, deny))

	for times, want := range rq {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://google.com?x=%d", times+1), nil)
		req.Header.Set("x", f.ToString(times+1))
		res, _ := app.Test(req)
		if want != res.StatusCode {
			t.Logf("%d != %d", want, res.StatusCode)
			t.Log(f.ToString(res.Body))
			t.Error(ErrTooManyRequests)
		}
	}

	// wait the interval, should be good for more
	time.Sleep(time.Second)
	for times, want := range rq {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://google.com?x=%d", times+1), nil)
		req.Header.Set("x", f.ToString(times+1))
		res, _ := app.Test(req)
		if want != res.StatusCode {
			t.Log(f.ToString(res.Body))
			t.Error(ErrTooManyRequests)
		}
	}
}
