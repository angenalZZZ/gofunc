package f

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// GbkToUtf8 convert encoding simplified chinese text.
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// ToUtf8 convert encode simplified to utf-8.
func ToUtf8(s []byte, encode string) ([]byte, error) {
	byteReader := bytes.NewReader(s)
	reader, err := charset.NewReaderLabel(encode, byteReader)
	if err != nil {
		return nil, err
	}
	if dst, err := ioutil.ReadAll(reader); err != nil {
		return nil, err
	} else {
		return dst, nil
	}
}
