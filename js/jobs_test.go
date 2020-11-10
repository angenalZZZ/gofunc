package js

import (
	"testing"

	"github.com/angenalZZZ/gofunc/log"
)

func TestNewJobs(t *testing.T) {
	r := NewRuntime(nil)
	defer func() { r.Clear() }()

	script := `
cron = [
    {
        name: "001",
        spec: "* * * * *", // every minutes
        func: function () {
            var item = { Time: new Date() };
            item.ActionName = 'some action';
            var res = $.q("post", "https://postman-echo.com/post", item, "url");
            log.debug(JSON.stringify(res));
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
		t.Log("jobs[0] init ok.")
		//t.Logf("jobs[0]: %+v", jobs[0])
		//t.Logf("jobs[0].FileIsMod: %t", jobs[0].FileIsMod())
		jobs[0].R, jobs[0].P = NewRuntime, &GoRuntimeParam{
			Log: log.InitConsole("15:04:05", false),
		}
		jobs[0].Run()
	}
}
