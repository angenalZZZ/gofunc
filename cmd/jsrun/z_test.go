package main

import "testing"

func Test(t *testing.T) {
	// Check Arguments And Init Config.
	checkArgs()

	// New Client Connect.
	natClientConnect()

	t.Logf("load js %q", scriptFile)

	// Run script test.
	run()
}
