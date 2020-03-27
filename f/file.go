package f

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CurrentPath gets compiled executable file absolute path.
func CurrentPath() string {
	path, _ := filepath.Abs(os.Args[0])
	return path
}

// CurrentDir gets compiled executable file directory.
func CurrentDir() string {
	return filepath.Dir(CurrentPath())
}

// RelativePath gets relative path.
func RelativePath(targetPath string) string {
	basePath, _ := filepath.Abs("./")
	rel, _ := filepath.Rel(basePath, targetPath)
	return strings.Replace(rel, `\`, `/`, -1)
}

// FileExists reports whether the named file or directory exists.
func FileExists(name string) (existed bool) {
	existed, _ = FileExist(name)
	return
}

// FileExist reports whether the named file or directory exists.
func FileExist(name string) (existed bool, isDir bool) {
	info, err := os.Stat(name)
	if err != nil {
		return !os.IsNotExist(err), false
	}
	return true, info.IsDir()
}

// SearchFile Search a file in paths.
// this is often used in search config file in /etc ~/
func SearchFile(filename string, paths ...string) (fullPath string) {
	for _, path := range paths {
		fullPath = filepath.Join(path, filename)
		existed, _ := FileExist(fullPath)
		if existed {
			return
		}
	}
	return
}

// MatchFile like command grep -E
// for example: MatchFile(`^hello`, "hello.txt")
// \n is striped while read
func MatchFile(patten string, filename string) (lines []string, err error) {
	re, err := regexp.Compile(patten)
	if err != nil {
		return
	}

	fd, err := os.Open(filename)
	if err != nil {
		return
	}
	lines = make([]string, 0)
	reader := bufio.NewReader(fd)
	prefix := ""
	isLongLine := false
	for {
		byteLine, isPrefix, er := reader.ReadLine()
		if er != nil && er != io.EOF {
			return nil, er
		}
		if er == io.EOF {
			break
		}
		line := string(byteLine)
		if isPrefix {
			prefix += line
			continue
		} else {
			isLongLine = true
		}

		line = prefix + line
		if isLongLine {
			prefix = ""
		}
		if re.MatchString(line) {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

// WalkDirs traverses the directory, return to the relative path.
// You can specify the suffix.
func WalkDirs(targetPath string, suffixes ...string) (dirList []string) {
	if !filepath.IsAbs(targetPath) {
		targetPath, _ = filepath.Abs(targetPath)
	}
	_ = filepath.Walk(targetPath, func(retPath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !f.IsDir() {
			return nil
		}
		if len(suffixes) == 0 {
			dirList = append(dirList, RelativePath(retPath))
			return nil
		}
		_path := RelativePath(retPath)
		for _, suffix := range suffixes {
			if strings.HasSuffix(_path, suffix) {
				dirList = append(dirList, _path)
			}
		}
		return nil
	})
	return
}

// FilepathSplitExt splits the filename into a pair (root, ext) such that root + ext == filename,
// and ext is empty or begins with a period and contains at most one period.
// Leading periods on the basename are ignored; splitext('.cshrc') returns ('', '.cshrc').
func FilepathSplitExt(filename string, slashInsensitive ...bool) (root, ext string) {
	insensitive := false
	if len(slashInsensitive) > 0 {
		insensitive = slashInsensitive[0]
	}
	if insensitive {
		filename = FilepathSlashInsensitive(filename)
	}
	for i := len(filename) - 1; i >= 0 && !os.IsPathSeparator(filename[i]); i-- {
		if filename[i] == '.' {
			return filename[:i], filename[i:]
		}
	}
	return filename, ""
}

// FilepathStem returns the stem of filename.
// Example:
//  FilepathStem("/root/dir/sub/file.ext") // output "file"
// NOTE:
//  If slashInsensitive is empty, default is false.
func FilepathStem(filename string, slashInsensitive ...bool) string {
	insensitive := false
	if len(slashInsensitive) > 0 {
		insensitive = slashInsensitive[0]
	}
	if insensitive {
		filename = FilepathSlashInsensitive(filename)
	}
	base := filepath.Base(filename)
	for i := len(base) - 1; i >= 0; i-- {
		if base[i] == '.' {
			return base[:i]
		}
	}
	return base
}

// FilepathSlashInsensitive ignore the difference between the slash and the backslash,
// and convert to the same as the current system.
func FilepathSlashInsensitive(path string) string {
	if filepath.Separator == '/' {
		return strings.Replace(path, "\\", "/", -1)
	}
	return strings.Replace(path, "/", "\\", -1)
}

// FilepathContains checks if the basePath path contains the subPaths.
func FilepathContains(basePath string, subPaths []string) error {
	basePath, err := filepath.Abs(basePath)
	if err != nil {
		return err
	}
	for _, p := range subPaths {
		p, err = filepath.Abs(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(basePath, p)
		if err != nil {
			return err
		}
		if strings.HasPrefix(rel, "..") {
			return fmt.Errorf("%s is not include %s", basePath, p)
		}
	}
	return nil
}

// FilepathAbsolute returns the absolute paths.
func FilepathAbsolute(paths []string) ([]string, error) {
	return StringsConvert(paths, func(p string) (string, error) {
		return filepath.Abs(p)
	})
}

// FilepathAbsoluteMap returns the absolute paths map.
func FilepathAbsoluteMap(paths []string) (map[string]string, error) {
	return StringsConvertMap(paths, func(p string) (string, error) {
		return filepath.Abs(p)
	})
}

// FilepathRelative returns the relative paths.
func FilepathRelative(basePath string, targetPaths []string) ([]string, error) {
	basePath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}
	return StringsConvert(targetPaths, func(p string) (string, error) {
		return filepathRelative(basePath, p)
	})
}

// FilepathRelativeMap returns the relative paths map.
func FilepathRelativeMap(basePath string, targetPaths []string) (map[string]string, error) {
	basePath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}
	return StringsConvertMap(targetPaths, func(p string) (string, error) {
		return filepathRelative(basePath, p)
	})
}

func filepathRelative(basePath, targetPath string) (string, error) {
	abs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(basePath, abs)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("%s is not include %s", basePath, abs)
	}
	return rel, nil
}

// FilepathDistinct removes the same path and return in the original order.
// If toAbs is true, return the result to absolute paths.
func FilepathDistinct(paths []string, toAbs bool) ([]string, error) {
	m := make(map[string]bool, len(paths))
	ret := make([]string, 0, len(paths))
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		if m[abs] {
			continue
		}
		m[abs] = true
		if toAbs {
			ret = append(ret, abs)
		} else {
			ret = append(ret, p)
		}
	}
	return ret, nil
}

// FilepathToSlash returns the result of replacing each separator character
// in path with a slash ('/') character. Multiple separators are
// replaced by multiple slashes.
func FilepathToSlash(paths []string) []string {
	ret, _ := StringsConvert(paths, func(p string) (string, error) {
		return filepath.ToSlash(p), nil
	})
	return ret
}

// FilepathFromSlash returns the result of replacing each slash ('/') character
// in path with a separator character. Multiple slashes are replaced
// by multiple separators.
func FilepathFromSlash(paths []string) []string {
	ret, _ := StringsConvert(paths, func(p string) (string, error) {
		return filepath.FromSlash(p), nil
	})
	return ret
}

// FilepathSame checks if the two paths are the same.
func FilepathSame(path1, path2 string) (bool, error) {
	if path1 == path2 {
		return true, nil
	}
	p1, err := filepath.Abs(path1)
	if err != nil {
		return false, err
	}
	p2, err := filepath.Abs(path2)
	if err != nil {
		return false, err
	}
	return p1 == p2, nil
}

// MkdirAll creates a directory named path,
// along with any necessary parents, and returns nil,
// or else returns an error.
// The permission bits perm (before umask) are used for all
// directories that MkdirAll creates.
// If path is already a directory, MkdirAll does nothing
// and returns nil.
// If perm is empty, default use 0755.
func MkdirAll(path string, perm ...os.FileMode) error {
	var fm os.FileMode = 0755
	if len(perm) > 0 {
		fm = perm[0]
	}
	return os.MkdirAll(path, fm)
}

// WriteFile writes file, and automatically creates the directory if necessary.
// NOTE:
//  If perm is empty, automatically determine the file permissions based on extension.
func WriteFile(filename string, data []byte, perm ...os.FileMode) error {
	filename = filepath.FromSlash(filename)
	err := MkdirAll(filepath.Dir(filename))
	if err != nil {
		return err
	}
	if len(perm) > 0 {
		return ioutil.WriteFile(filename, data, perm[0])
	}
	var ext string
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		ext = filename[idx:]
	}
	switch ext {
	case ".sh", ".py", ".rb", ".bat", ".com", ".vbs", ".htm", ".run", ".App", ".exe", ".reg":
		return ioutil.WriteFile(filename, data, 0755)
	default:
		return ioutil.WriteFile(filename, data, 0644)
	}
}

// RewriteFile rewrites the file.
func RewriteFile(filename string, fn func(content []byte) (newContent []byte, err error)) error {
	f, err := os.OpenFile(filename, os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	newContent, err := fn(content)
	if err != nil {
		return err
	}
	if bytes.Equal(content, newContent) {
		return nil
	}
	_, _ = f.Seek(0, 0)
	_ = f.Truncate(0)
	_, err = f.Write(newContent)
	return err
}

// RewriteToFile rewrites the file to newFilename.
// If newFilename already exists and is not a directory, replaces it.
func RewriteToFile(filename, newFilename string, fn func(content []byte) (newContent []byte, err error)) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	cnt, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	newContent, err := fn(cnt)
	if err != nil {
		return err
	}
	return WriteFile(newFilename, newContent, info.Mode())
}

// ReplaceFile replaces the bytes selected by [start, end] with the new content.
func ReplaceFile(filename string, start, end int, newContent string) error {
	if start < 0 || (end >= 0 && start > end) {
		return nil
	}
	return RewriteFile(filename, func(content []byte) ([]byte, error) {
		if end < 0 || end > len(content) {
			end = len(content)
		}
		if start > end {
			start = end
		}
		return bytes.Replace(content, content[start:end], StringToBytes(newContent), 1), nil
	})
}
