package f

import (
	"testing"
)

func TestExecCommandOutputScanner(t *testing.T) {
	name := "guid.exe"
	if err := ExecCommandOutputScanner(name, []string{"-n=3"}, func(line []byte) bool {
		t.Logf("%s\n", line)
		return false
	}); err != nil {
		t.Fatal(err)
	}
}
