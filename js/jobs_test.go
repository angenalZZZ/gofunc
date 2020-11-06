package js

import (
	"testing"

	"github.com/dop251/goja"
)

func TestNewJobs(t *testing.T) {
	r := goja.New()
	defer func() { r.ClearInterrupt() }()
	//Console(r)

	script := `
cron = [
    {
        name: "001",
        spec: "* * * * *", // every minutes
        func: function () {
            var item = { Time: new Date() };
            item.ActionName = 'some action';
            var res = $.q("post", "https://postman-echo.com/post", item, "url");
            dump(res);
        }
    },
];
`

	if jobs, err := NewJobs(r, script, "cron", ""); err != nil {
		t.Fatal(err)
	} else {
		if err = jobs[0].Init(); err != nil {
			t.Fatal(err)
		}
		t.Logf("jobs[0]: %+v", jobs[0])
		t.Logf("jobs[0].FileIsMod: %t", jobs[0].FileIsMod())
	}
}
