package f

import "errors"

// Must not error, or panic.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustBytes bytes length is these numbers, or panic.
func MustBytes(p []byte, n ...int) {
	if n != nil && len(n) > 0 {
		if p == nil {
			panic(errors.New("wrong empty bytes"))
		}
		ok := false
		l := len(p)
		for _, i := range n {
			if i == l {
				ok = true
				break
			}
		}
		if ok == false {
			panic(errors.New("wrong bytes length"))
		}
	}
}
