package f_test

import (
	"github.com/angenalZZZ/gofunc/f"
	"os"
	"testing"
)

func TestZipCompress(t *testing.T) {
	// Zip Compress
	destination, sources := "../test/rsa.zip", []string{
		"../test/rsa",
	}

	if f.FileExists(destination) {
		t.Logf(" file exists: %s , is zip file: %t\n", destination, f.IsZipFile(destination))
	}

	if err := f.ZipCompress(sources, destination, true, false); err != nil {
		t.Fatal(err)
	}

	// Zip Decompress
	reader, err := f.ZipOpenReader(destination)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = reader.Close()
		_ = os.Remove(destination)
	}()

	if err := f.ZipDecompress(&reader.Reader, "../test/"); err != nil {
		t.Fatal(err)
	}
}
