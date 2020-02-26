package gofunc

// Must not error, or panic.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}
