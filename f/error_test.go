package f_test

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/f"
	"testing"
)

func TestErrorsToString(t *testing.T) {
	t.Parallel()
	customErr := &f.Error{Name: "Custom Error Name", Err: fmt.Errorf("stdlib error")}
	customErrWithCustomErrorMessage := &f.Error{Name: "Custom Error Name 2", Err: fmt.Errorf("Bad stuff happened"), CustomErrorMessageExists: true}

	var tests = []struct {
		param1   f.Errors
		expected string
	}{
		{f.Errors{}, ""},
		{f.Errors{fmt.Errorf("Error 1")}, "Error 1"},
		{f.Errors{fmt.Errorf("Error 1"), fmt.Errorf("Error 2")}, "Error 1;Error 2"},
		{f.Errors{customErr, fmt.Errorf("Error 2")}, "Custom Error Name: stdlib error;Error 2"},
		{f.Errors{fmt.Errorf("Error 123"), customErrWithCustomErrorMessage}, "Error 123;Bad stuff happened"},
	}
	for _, test := range tests {
		actual := test.param1.Error()
		if actual != test.expected {
			t.Errorf("Expected Error() to return '%v', got '%v'", test.expected, actual)
		}
	}
}
