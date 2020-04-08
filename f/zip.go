package f

import (
	"bytes"
	"fmt"
	"github.com/klauspost/compress/zip"
	gzip "github.com/klauspost/pgzip"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ZipOpenReader open a zip reader.
func ZipOpenReader(zipFile string) (*zip.ReadCloser, error) {
	return zip.OpenReader(zipFile)
}

// ZipNewReader gets a zip reader.
func ZipNewReader(zipFile string) (*zip.Reader, error) {
	file, err := os.Open(zipFile)
	if err != nil {
		return nil, err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	reader, err := zip.NewReader(file, fileInfo.Size())
	if err != nil {
		return nil, err
	}
	return reader, nil
}

// GzipNewReader gets a gzip reader.
func GzipNewReader(zipFile string) (*gzip.Reader, error) {
	file, err := os.Open(zipFile)
	if err != nil {
		return nil, err
	}
	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

// ZipFindAll returns the cleaned path of every file in the supplied zip reader.
func ZipFindAll(reader *zip.Reader) ([]string, error) {
	return ZipFind(reader, "*")
}

// ZipFind returns the cleaned path of every file in the supplied zip reader whose
// base name matches the supplied pattern, which is interpreted as in path.Match.
func ZipFind(reader *zip.Reader, pattern string) ([]string, error) {
	// path.Match will only return an error if the pattern is not
	// valid (*and* the supplied name is not empty, hence "check").
	if _, err := path.Match(pattern, "check"); err != nil {
		return nil, err
	}
	var matches []string
	for _, zipFile := range reader.File {
		cleanPath := path.Clean(zipFile.Name)
		baseName := path.Base(cleanPath)
		if match, _ := path.Match(pattern, baseName); match {
			matches = append(matches, cleanPath)
		}
	}
	return matches, nil
}

// ZipExtractAll extracts the supplied zip reader to the target path, overwriting
// existing files and directories only where necessary.
func ZipExtractAll(reader *zip.Reader, targetRoot string) error {
	return ZipExtract(reader, targetRoot, "")
}

// ZipExtract extracts files from the supplied zip reader, from the (internal, slash-
// separated) source path into the (external, OS-specific) target path. If the
// source path does not reference a directory, the referenced file will be written
// directly to the target path.
func ZipExtract(reader *zip.Reader, targetRoot, sourceRoot string) error {
	sourceRoot = path.Clean(sourceRoot)
	if sourceRoot == "." {
		sourceRoot = ""
	}
	if !IsSanePath(sourceRoot) {
		return fmt.Errorf("cannot extract files rooted at %q", sourceRoot)
	}
	extractor := zipExtractor{targetRoot, sourceRoot}
	for _, zipFile := range reader.File {
		if err := extractor.extract(zipFile); err != nil {
			cleanName := path.Clean(zipFile.Name)
			return fmt.Errorf("cannot extract %q: %v", cleanName, err)
		}
	}
	return nil
}

type zipExtractor struct {
	targetRoot string
	sourceRoot string
}

// targetPath returns the target path for a given zip file and whether
// it should be extracted.
func (x zipExtractor) targetPath(zipFile *zip.File) (string, bool) {
	cleanPath := path.Clean(zipFile.Name)
	if cleanPath == x.sourceRoot {
		return x.targetRoot, true
	}
	if x.sourceRoot != "" {
		mustPrefix := x.sourceRoot + "/"
		if !strings.HasPrefix(cleanPath, mustPrefix) {
			return "", false
		}
		cleanPath = cleanPath[len(mustPrefix):]
	}
	return filepath.Join(x.targetRoot, filepath.FromSlash(cleanPath)), true
}

func (x zipExtractor) extract(zipFile *zip.File) error {
	targetPath, ok := x.targetPath(zipFile)
	if !ok {
		return nil
	}
	parentPath := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentPath, 0777); err != nil {
		return err
	}
	mode := zipFile.Mode()
	modePerm := mode & os.ModePerm
	modeType := mode & os.ModeType
	switch modeType {
	case os.ModeDir:
		return x.writeDir(targetPath, modePerm)
	case os.ModeSymlink:
		return x.writeSymlink(targetPath, zipFile)
	case 0:
		return x.writeFile(targetPath, zipFile, modePerm)
	}
	return fmt.Errorf("unknown file type %d", modeType)
}

func (x zipExtractor) writeDir(targetPath string, modePerm os.FileMode) error {
	fileInfo, err := os.Lstat(targetPath)
	switch {
	case err == nil:
		mode := fileInfo.Mode()
		if mode.IsDir() {
			if mode&os.ModePerm != modePerm {
				return os.Chmod(targetPath, modePerm)
			}
			return nil
		}
		fallthrough
	case !os.IsNotExist(err):
		if err := os.RemoveAll(targetPath); err != nil {
			return err
		}
	}
	return os.MkdirAll(targetPath, modePerm)
}

func (x zipExtractor) writeFile(targetPath string, zipFile *zip.File, modePerm os.FileMode) error {
	if _, err := os.Lstat(targetPath); !os.IsNotExist(err) {
		if err := os.RemoveAll(targetPath); err != nil {
			return err
		}
	}
	writer, err := os.OpenFile(targetPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, modePerm)
	if err != nil {
		return err
	}
	defer writer.Close()

	if err := zipCopyTo(writer, zipFile); err != nil {
		return err
	}

	if err := writer.Sync(); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}

func (x zipExtractor) writeSymlink(targetPath string, zipFile *zip.File) error {
	symlinkTarget, err := x.checkSymlink(targetPath, zipFile)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(targetPath); !os.IsNotExist(err) {
		if err := os.RemoveAll(targetPath); err != nil {
			return err
		}
	}
	return os.Symlink(symlinkTarget, targetPath)
}

func (x zipExtractor) checkSymlink(targetPath string, zipFile *zip.File) (string, error) {
	var buffer bytes.Buffer
	if err := zipCopyTo(&buffer, zipFile); err != nil {
		return "", err
	}
	symlinkTarget := buffer.String()
	if filepath.IsAbs(symlinkTarget) {
		return "", fmt.Errorf("symlink %q is absolute", symlinkTarget)
	}
	finalPath := filepath.Join(filepath.Dir(targetPath), symlinkTarget)
	relativePath, err := filepath.Rel(x.targetRoot, finalPath)
	if err != nil {
		// Not tested, because I don't know how to trigger this condition.
		return "", fmt.Errorf("symlink %q not comprehensible", symlinkTarget)
	}
	if !IsSanePath(relativePath) {
		return "", fmt.Errorf("symlink %q leads out of scope", symlinkTarget)
	}
	return symlinkTarget, nil
}

func zipCopyTo(writer io.Writer, zipFile *zip.File) error {
	reader, err := zipFile.Open()
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, reader)
	_ = reader.Close()
	return err
}
