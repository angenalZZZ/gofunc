package f

import "io/ioutil"

// ReadFile reads the file named by filename and returns the contents.
// A successful call returns err == nil, not err == EOF. Because ReadFile
// reads the whole file, it does not treat an EOF from Read as an error
// to be reported.
func ReadFile(filename string) ([]byte, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	// windows file prefix bytes problem
	if len(b) > 3 && b[0] == 239 && b[1] == 187 && b[2] == 191 {
		return b[3:], nil
	}
	return b, nil
}
