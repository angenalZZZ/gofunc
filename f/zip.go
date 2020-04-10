package f

import (
	"bytes"
	"fmt"
	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zip"
	"github.com/klauspost/pgzip"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// GzipCompress reads in, compresses it, and writes it to out.
func GzipCompress(in io.Reader, out io.Writer, singleThreaded bool, compressionLevels ...int) error {
	compressionLevel := gzip.DefaultCompression
	if len(compressionLevels) == 1 {
		compressionLevel = compressionLevels[0]
	}
	var w io.WriteCloser
	var err error
	if singleThreaded {
		w, err = gzip.NewWriterLevel(out, compressionLevel)
	} else {
		w, err = pgzip.NewWriterLevel(out, compressionLevel)
	}
	if err != nil {
		return err
	}
	_, err = io.Copy(w, in)
	_ = w.Close()
	return err
}

// GzipDecompress reads in, decompresses it, and writes it to out.
func GzipDecompress(in io.Reader, out io.Writer, singleThreaded bool) error {
	var r io.ReadCloser
	var err error
	if singleThreaded {
		r, err = gzip.NewReader(in)
	} else {
		r, err = pgzip.NewReader(in)
	}
	if err != nil {
		return err
	}
	_, err = io.Copy(out, r)
	_ = r.Close()
	return err
}

// IsZipFile check is zip file.
func IsZipFile(file string) bool {
	f, err := os.Open(file)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 4)
	if n, err := f.Read(buf); err != nil || n < 4 {
		return false
	}

	return bytes.Equal(buf, []byte("PK\x03\x04"))
}

// ZipCompress creates a .zip file at destination containing
// the files listed in sources. The destination must end
// with ".zip". zipFileInfo paths can be those of regular files
// or directories. Regular files are stored at the 'root'
// of the archive, and directories are recursively added.
func ZipCompress(sources []string, destination string, overwriteExisting, implicitTopLevelFolder bool, compressionLevels ...int) error {
	if !strings.HasSuffix(destination, ".zip") {
		return fmt.Errorf("filename must have a .zip extension")
	}
	if !overwriteExisting && FileExists(destination) {
		return fmt.Errorf("file already exists: %s", destination)
	}
	compressionLevel := flate.DefaultCompression
	if len(compressionLevels) == 1 {
		compressionLevel = compressionLevels[0]
	}
	// make the folder to contain the resulting archive
	// if it does not already exist
	destDir := filepath.Dir(destination)
	if !FileExists(destDir) {
		err := MkdirAll(destDir)
		if err != nil {
			return fmt.Errorf("making folder for destination: %v", err)
		}
	}

	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("creating %s: %v", destination, err)
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	if compressionLevel != flate.DefaultCompression {
		zw.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(out, compressionLevel)
		})
	}
	defer zw.Close()

	var topLevelFolder string
	if implicitTopLevelFolder && FileExistMultipleTopLevels(sources) {
		topLevelFolder = FolderNameFromFileName(destination)
	}

	for _, source := range sources {
		err := zipWriteWalk(source, topLevelFolder, destination, zw)
		if err != nil {
			return fmt.Errorf("walking %s: %v", source, err)
		}
	}

	return nil
}

func zipWriteWalk(source, topLevelFolder, destination string, zw *zip.Writer) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("%s: stat: %v", source, err)
	}
	destAbs, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("%s: getting absolute path of destination %s: %v", source, destination, err)
	}

	return filepath.Walk(source, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("traversing %s: %v", fpath, err)
		}
		if info == nil {
			return fmt.Errorf("%s: no file info", fpath)
		}

		// make sure we do not copy the output file into the output
		// file; that results in an infinite loop and disk exhaustion!
		fpathAbs, err := filepath.Abs(fpath)
		if err != nil {
			return fmt.Errorf("%s: getting absolute path: %v", fpath, err)
		}
		if FileWithin(fpathAbs, destAbs) {
			return nil
		}

		// build the name to be used within the archive
		nameInArchive, err := MakeNameInArchive(sourceInfo, source, topLevelFolder, fpath)
		if err != nil {
			return err
		}

		var file io.ReadCloser
		if info.Mode().IsRegular() {
			file, err = os.Open(fpath)
			if err != nil {
				return fmt.Errorf("%s: opening: %v", fpath, err)
			}
			defer file.Close()
		}
		err = zipWrite(zipFileInfo{
			FileInfo: zipFileCustomInfo{
				FileInfo:   info,
				CustomName: nameInArchive,
			},
			ReadCloser: file,
		}, zw)
		if err != nil {
			return fmt.Errorf("%s: writing: %s", fpath, err)
		}

		return nil
	})
}

// Write writes f to z, which must have been opened for writing first.
func zipWrite(f zipFileInfo, zw *zip.Writer) error {
	if f.FileInfo.Name() == "" {
		return fmt.Errorf("missing file name")
	}

	header, err := zip.FileInfoHeader(f)
	if err != nil {
		return fmt.Errorf("%s: getting header: %v", f.Name(), err)
	}

	if f.IsDir() {
		header.Name += "/" // required - strangely no mention of this in zip spec? but is in godoc...
		header.Method = zip.Store
	} else {
		ext := strings.ToLower(path.Ext(header.Name))
		if _, ok := CompressedFormats[ext]; ok {
			header.Method = zip.Store
		} else {
			header.Method = zip.Deflate
		}
	}

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("%s: making header: %v", f.Name(), err)
	}

	return zipWriteFile(f, writer)
}

func zipWriteFile(f zipFileInfo, writer io.Writer) error {
	if f.IsDir() {
		return nil // directories have no contents
	}
	if FileIsSymlink(f) {
		// file body for symlinks is the symlink target
		linkTarget, err := os.Readlink(f.Name())
		if err != nil {
			return fmt.Errorf("%s: readlink: %v", f.Name(), err)
		}
		_, err = writer.Write([]byte(filepath.ToSlash(linkTarget)))
		if err != nil {
			return fmt.Errorf("%s: writing symlink target: %v", f.Name(), err)
		}
		return nil
	}

	if f.ReadCloser == nil {
		return fmt.Errorf("%s: no way to read file contents", f.Name())
	}
	_, err := io.Copy(writer, f)
	if err != nil {
		return fmt.Errorf("%s: copying contents: %v", f.Name(), err)
	}

	return nil
}

// zipFileInfo provides methods for accessing information about
// or contents of a file within an archive.
type zipFileInfo struct {
	os.FileInfo

	// The original header info; depends on
	// type of archive -- could be nil, too.
	Header interface{}

	// Allow the file contents to be read (and closed)
	io.ReadCloser
}

// zipFileCustomInfo is an os.zipFileCustomInfo but optionally with
// a custom name, useful if dealing with files that
// are not actual files on disk, or which have a
// different name in an archive than on disk.
type zipFileCustomInfo struct {
	os.FileInfo
	CustomName string
}

// Name returns fi.CustomName if not empty;
// otherwise it returns fi.zipFileCustomInfo.Name().
func (fi zipFileCustomInfo) Name() string {
	if fi.CustomName != "" {
		return fi.CustomName
	}
	return fi.FileInfo.Name()
}

// ZipOpenReader open a zip reader.
var ZipOpenReader = zip.OpenReader

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

// ZipFind returns the cleaned path of every file in the supplied zip reader whose
// base name matches the supplied pattern, which is interpreted as in path.Match.
func ZipFind(reader *zip.Reader, patterns ...string) ([]string, error) {
	// path.Match will only return an error if the pattern is not
	// valid (*and* the supplied name is not empty, hence "check").
	pattern := "*"
	if len(patterns) == 1 {
		pattern = patterns[0]
	}
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

// ZipDecompress extracts files from the supplied zip reader, from the (internal, slash-
// separated) source path into the (external, OS-specific) target path. If the
// source path does not reference a directory, the referenced file will be written
// directly to the target path.
func ZipDecompress(reader *zip.Reader, targetRoot string, sourceRoot ...string) error {
	source := ""
	if len(sourceRoot) == 1 {
		source = sourceRoot[0]
	}
	source = path.Clean(source)
	if source == "." {
		source = ""
	}
	if !IsSanePath(source) {
		return fmt.Errorf("cannot extract files rooted at %q", source)
	}
	extractor := zipExtractor{targetRoot, source}
	for _, zipFile := range reader.File {
		if err := extractor.extract(zipFile); err != nil {
			cleanName := path.Clean(zipFile.Name)
			return fmt.Errorf("cannot extract %q: %v", cleanName, err)
		}
	}
	return nil
}

// zipExtractor extracts files from the supplied zip path.
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

// CompressedFormats is a (non-exhaustive) set of lowerCased
// file extensions for formats that are typically already
// compressed. Compressing files that are already compressed
// is inefficient, so use this set of extension to avoid that.
var CompressedFormats = map[string]struct{}{
	".7z":   {},
	".avi":  {},
	".br":   {},
	".bz2":  {},
	".cab":  {},
	".docx": {},
	".gif":  {},
	".gz":   {},
	".jar":  {},
	".jpeg": {},
	".jpg":  {},
	".lz":   {},
	".lz4":  {},
	".lzma": {},
	".m4v":  {},
	".mov":  {},
	".mp3":  {},
	".mp4":  {},
	".mpeg": {},
	".mpg":  {},
	".png":  {},
	".pptx": {},
	".rar":  {},
	".sz":   {},
	".tbz2": {},
	".tgz":  {},
	".tsz":  {},
	".txz":  {},
	".xlsx": {},
	".xz":   {},
	".zip":  {},
	".zipx": {},
}
