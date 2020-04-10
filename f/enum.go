package f

// Times returns a slice of n 0-sized elements, suitable for ranging over.
//
// For example:
//
//    for i := range Times(10) {
//        fmt.Println(i)
//    }
//    Times(10, func(i int) {
//    	fmt.Print(i)
//    })
//
// ... will print 0 to 9, inclusive.
//
// It does not cause any allocations.
func Times(n int, fn ...func(i int)) (s []struct{}) {
	s = make([]struct{}, n)
	for _, f := range fn {
		for i := range s {
			f(i)
		}
	}
	return
}

// TimesRepeat create times.
func TimesRepeat(times int, value interface{}) []interface{} {
	q := make([]interface{}, times)
	for i := range q {
		q[i] = value
	}
	return q
}

// TimesRepeatAppend append times.
func TimesRepeatAppend(slice []interface{}, times int, value interface{}) {
	if slice == nil {
		slice = make([]interface{}, 0, times)
	}
	q := TimesRepeat(times, value)
	slice = append(slice, q...)
	return
}
