package f

import "testing"

func TestZipCompress(t *testing.T) {
	// Zip Compress
	destination, sources := "../test/rsa.zip", []string{
		"../test/rsa",
	}
	if err := ZipCompress(sources, destination, true, false); err != nil {
		t.Fatal(err)
	}

	// Zip Decompress
	reader, err := ZipOpenReader(destination)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if err := ZipDecompress(&reader.Reader, "../test/"); err != nil {
		t.Fatal(err)
	}
}
