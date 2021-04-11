package snowid

import "fmt"

// IdGeneratorException .
type IdGeneratorException struct {
	message string
	error   error
}

// IdGeneratorException .
func (e IdGeneratorException) IdGeneratorException(message ...interface{}) {
	fmt.Println(message...)
}

// Error .
func (e IdGeneratorException) Error(err error) string {
	e.message = err.Error()
	e.error = err
	return e.message
}
