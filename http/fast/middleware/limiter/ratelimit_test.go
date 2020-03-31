package limiter

import (
	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/http/fast"
	"net/http"
	"testing"
)

func TestRateLimit(t *testing.T) {
	app := fast.New()

	times := 300
	rl := NewRateLimiter(int64(times), "too many requests")
	rq := make([]int, times)
	for i, _ := range rq {
		rq[i] = http.StatusOK
	}
	rq = append(rq, rl.StatusCode)
	t.Logf("%d : %v", len(rq), rq)

	allow, _ := func(c *fast.Ctx) {
		c.SendString(c.OriginalURL())
	}, func(c *fast.Ctx) {
		c.Status(rl.StatusCode).SendString(rl.Message + " $x=" + c.GetHeader("x"))
	}

	app.Get("/", rl.Wrap(allow, nil))

	for times, want := range rq {
		req, _ := http.NewRequest("GET", "http://google.com?x="+f.ToString(times+1), nil)
		req.Header.Set("x", f.ToString(times+1))
		res, _ := app.Test(req)
		if want != res.StatusCode {
			t.Logf("want != res.StatusCode > %d != %d", want, res.StatusCode)
			t.Log(f.ToString(res.Body))
			t.Error(ErrTooManyRequests)
		} else {
			//t.Log(f.ToString(res.Body))
		}
	}

	// wait the interval, should be good for 2 more
	//time.Sleep(time.Second)
	//for _, want := range rq {
	//	req, _ := http.NewRequest("GET", "http://google.com", nil)
	//	req.Header.Set("x", f.ToString(times+1))
	//	res, _ := app.Test(req)
	//	if want != res.StatusCode {
	//		t.Log(f.ToString(res.Body))
	//		t.Error(ErrTooManyRequests)
	//	} else {
	//		//t.Log(f.ToString(res.Body))
	//	}
	//}
}
