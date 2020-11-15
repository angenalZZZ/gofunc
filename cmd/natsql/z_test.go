package main

import (
	"testing"

	"github.com/angenalZZZ/gofunc/data"
)

func Test(t *testing.T) {
	isTest = true
	jsonFile = "data.json"
	cacheDir = data.CodeDir("cmd/natsql/test-")

	// Check Arguments And Init Config.
	checkArgs()

	t.Logf("load js %q", scriptFile)

	// Run script test.
	runTest()
}
