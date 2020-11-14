package main

import (
	"os"
	"testing"
)

func Test(t *testing.T) {
	isTest = true
	jsonFile = "data.json"
	cacheDir = os.Getenv("GOPATH") + `\src\github.com\angenalZZZ\gofunc\cmd\natsql\test-`

	// Check Arguments And Init Config.
	checkArgs()

	t.Logf("load js %q", scriptFile)

	// Run script test.
	runTest()
}
