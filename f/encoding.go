package f

import (
	"bytes"
	"errors"
	"io"
	"os"
)

// Encoding is type alias for detected UTF encoding.
type Encoding int

// Constants to identify detected UTF encodings.
const (
	// Unknown encoding, returned when no BOM was detected
	UnknownEncoding Encoding = iota

	// UTF8, BOM bytes: EF BB BF
	UTF8

	// UTF-16, big-endian, BOM bytes: FE FF
	UTF16BigEndian

	// UTF-16, little-endian, BOM bytes: FF FE
	UTF16LittleEndian

	// UTF-32, big-endian, BOM bytes: 00 00 FE FF
	UTF32BigEndian

	// UTF-32, little-endian, BOM bytes: FF FE 00 00
	UTF32LittleEndian
)

const maxConsecutiveEmptyReads = 100

// String returns a user-friendly string representation of the encoding. Satisfies fmt.Stringer interface.
func (e Encoding) String() string {
	switch e {
	case UTF8:
		return "UTF8"
	case UTF16BigEndian:
		return "UTF16BigEndian"
	case UTF16LittleEndian:
		return "UTF16LittleEndian"
	case UTF32BigEndian:
		return "UTF32BigEndian"
	case UTF32LittleEndian:
		return "UTF32LittleEndian"
	default:
		return "UnknownEncoding"
	}
}

// ReadFile reads the file named by filename and returns the contents.
// File Reader which automatically detects BOM (Unicode Byte Order Mark) and removes it as necessary.
// A successful call returns err == nil, not err == EOF. Because ReadFile
// reads the whole file, it does not treat an EOF from Read as an error
// to be reported.
func ReadFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	// It's a good but not certain bet that FileInfo will tell us exactly how much to
	// read, so let's try it but be prepared for the answer to be wrong.
	var n int64 = bytes.MinRead

	if fi, err := f.Stat(); err == nil {
		// As initial capacity for readAll, use Size + a little extra in case Size
		// is zero, and to avoid another allocation after Read has filled the
		// buffer. The readAll call will read into its allocated internal buffer
		// cheaply. If the size was wrong, we'll either waste some space off the end
		// or reallocate as needed, but in the overwhelmingly common case we'll get
		// it just right.
		if size := fi.Size() + bytes.MinRead; size > n {
			n = size
		}
	}

	var buf bytes.Buffer
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	if int64(int(n)) == n {
		buf.Grow(int(n))
	}

	// Automatically detects BOM and removes it as necessary.
	r, _ := SkipBOM(f)
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err

	//b, err := ioutil.ReadAll(f)
	//if err != nil {
	//	return nil, err
	//}
	//// skip BOM
	//if len(b) > 3 && b[0] == 239 && b[1] == 187 && b[2] == 191 {
	//	return b[3:], nil
	//}
	//return b, nil
}

// ReadFileAndTrimSpace reads the file and trim head-tail space contents.
func ReadFileAndTrimSpace(filename string) ([]byte, error) {
	buf, err := ReadFile(filename)
	if err != nil {
		return nil, err
	}
	buf = bytes.TrimSpace(buf)
	return buf, nil
}

// ReadFileEncoding reads the file and returns detected encoding.
func ReadFileEncoding(filename string) Encoding {
	f, err := os.Open(filename)
	if err != nil {
		return UnknownEncoding
	}
	if enc, _, err := detectUtf(f); err == nil {
		return enc
	}
	return UnknownEncoding
}

// SkipBOM creates Reader which automatically detects BOM (Unicode Byte Order Mark) and removes it as necessary.
// It also returns the encoding detected by the BOM.
// If the detected encoding is not needed, you can call the SkipOnly function.
func SkipBOM(rd io.Reader) (*Reader, Encoding) {
	// Is it already a Reader?
	b, ok := rd.(*Reader)
	if ok {
		return b, UnknownEncoding
	}

	enc, left, err := detectUtf(rd)
	return &Reader{
		rd:  rd,
		buf: left,
		err: err,
	}, enc
}

// Reader implements automatic BOM (Unicode Byte Order Mark) checking and
// removing as necessary for an io.Reader object.
type Reader struct {
	rd  io.Reader // reader provided by the client
	buf []byte    // buffered data
	err error     // last error
}

// Read is an implementation of io.Reader interface.
// The bytes are taken from the underlying Reader, but it checks for BOMs, removing them as necessary.
func (r *Reader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if r.buf == nil {
		if r.err != nil {
			return 0, r.readErr()
		}

		return r.rd.Read(p)
	}

	// copy as much as we can
	n = copy(p, r.buf)
	r.buf = nilIfEmpty(r.buf[n:])
	return n, nil
}

func (r *Reader) readErr() error {
	err := r.err
	r.err = nil
	return err
}

var errNegativeRead = errors.New("utf-bom: reader returned negative count from read")

func detectUtf(rd io.Reader) (enc Encoding, buf []byte, err error) {
	buf, err = readBOM(rd)

	if len(buf) >= 4 {
		if isUTF32BigEndianBOM4(buf) {
			return UTF32BigEndian, nilIfEmpty(buf[4:]), err
		}
		if isUTF32LittleEndianBOM4(buf) {
			return UTF32LittleEndian, nilIfEmpty(buf[4:]), err
		}
	}

	if len(buf) > 2 && isUTF8BOM3(buf) {
		return UTF8, nilIfEmpty(buf[3:]), err
	}

	if (err != nil && err != io.EOF) || (len(buf) < 2) {
		return UnknownEncoding, nilIfEmpty(buf), err
	}

	if isUTF16BigEndianBOM2(buf) {
		return UTF16BigEndian, nilIfEmpty(buf[2:]), err
	}
	if isUTF16LittleEndianBOM2(buf) {
		return UTF16LittleEndian, nilIfEmpty(buf[2:]), err
	}

	return UnknownEncoding, nilIfEmpty(buf), err
}

func readBOM(rd io.Reader) (buf []byte, err error) {
	const maxBOMSize = 4
	var bom [maxBOMSize]byte // used to read BOM

	// read as many bytes as possible
	for nEmpty, n := 0, 0; err == nil && len(buf) < maxBOMSize; buf = bom[:len(buf)+n] {
		if n, err = rd.Read(bom[len(buf):]); n < 0 {
			panic(errNegativeRead)
		}
		if n > 0 {
			nEmpty = 0
		} else {
			nEmpty++
			if nEmpty >= maxConsecutiveEmptyReads {
				err = io.ErrNoProgress
			}
		}
	}
	return
}

func isUTF32BigEndianBOM4(buf []byte) bool {
	return buf[0] == 0x00 && buf[1] == 0x00 && buf[2] == 0xFE && buf[3] == 0xFF
}

func isUTF32LittleEndianBOM4(buf []byte) bool {
	return buf[0] == 0xFF && buf[1] == 0xFE && buf[2] == 0x00 && buf[3] == 0x00
}

func isUTF8BOM3(buf []byte) bool {
	return buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF
}

func isUTF16BigEndianBOM2(buf []byte) bool {
	return buf[0] == 0xFE && buf[1] == 0xFF
}

func isUTF16LittleEndianBOM2(buf []byte) bool {
	return buf[0] == 0xFF && buf[1] == 0xFE
}

func nilIfEmpty(buf []byte) (res []byte) {
	if len(buf) > 0 {
		res = buf
	}
	return
}
