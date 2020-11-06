package main

import "testing"

func Test(t *testing.T) {
	isTest = true

	// Check Arguments And Init Config.
	checkArgs()

	// New Client Connect.
	clientConnect()

	t.Logf("load js %q", scriptFile)

	// Run script test.
	runTest()
}
