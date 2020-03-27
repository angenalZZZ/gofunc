package f

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

// CopyInterface for delegating copy process to type
type CopyInterface interface {
	Copy() interface{}
}

// CloneInterface for delegating copy process to type
type CloneInterface interface {
	Clone() interface{}
}

// Copy create a deep copy of whatever is passed to it and returns the copy
// in an interface{}.  The returned value will need to be asserted to the correct type.
func Clone(src interface{}) interface{} {
	if src == nil {
		return nil
	}

	// Make the interface a reflect.Value
	original := reflect.ValueOf(src)

	// Make a copy of the same type as the original.
	cpy := reflect.New(original.Type()).Elem()

	// Recursively copy the original.
	copyRecursive(original, cpy)

	// Return the copy as an interface.
	return cpy.Interface()
}

// copyRecursive does the actual copying of the interface. It currently has
// limited support for what it can handle. Add as needed.
func copyRecursive(original, cpy reflect.Value) {
	// check for implement CloneInterface
	if original.CanInterface() {
		if copier, ok := original.Interface().(CopyInterface); ok {
			cpy.Set(reflect.ValueOf(copier.Copy()))
			return
		}
		if copier, ok := original.Interface().(CloneInterface); ok {
			cpy.Set(reflect.ValueOf(copier.Clone()))
			return
		}
	}

	// handle according to original's Kind
	switch original.Kind() {
	case reflect.Ptr:
		// Get the actual value being pointed to.
		originalValue := original.Elem()

		// if  it isn't valid, return.
		if !originalValue.IsValid() {
			return
		}
		cpy.Set(reflect.New(originalValue.Type()))
		copyRecursive(originalValue, cpy.Elem())

	case reflect.Interface:
		// If this is a nil, don't do anything
		if original.IsNil() {
			return
		}
		// Get the value for the interface, not the pointer.
		originalValue := original.Elem()

		// Get the value by calling Elem().
		copyValue := reflect.New(originalValue.Type()).Elem()
		copyRecursive(originalValue, copyValue)
		cpy.Set(copyValue)

	case reflect.Struct:
		t, ok := original.Interface().(time.Time)
		if ok {
			cpy.Set(reflect.ValueOf(t))
			return
		}
		// Go through each field of the struct and copy it.
		for i := 0; i < original.NumField(); i++ {
			// The Type's StructField for a given field is checked to see if StructField.PkgPath
			// is set to determine if the field is exported or not because CanSet() returns false
			// for settable fields.  I'm not sure why.
			if original.Type().Field(i).PkgPath != "" {
				continue
			}
			copyRecursive(original.Field(i), cpy.Field(i))
		}

	case reflect.Slice:
		if original.IsNil() {
			return
		}
		// Make a new slice and copy each element.
		cpy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i++ {
			copyRecursive(original.Index(i), cpy.Index(i))
		}

	case reflect.Map:
		if original.IsNil() {
			return
		}
		cpy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			copyRecursive(originalValue, copyValue)
			copyKey := Clone(key.Interface())
			cpy.SetMapIndex(reflect.ValueOf(copyKey), copyValue)
		}

	default:
		cpy.Set(original)
	}
}

// Copy recursively copies the file, directory or symbolic link at src
// to dst. The destination must not exist. Symbolic links are not
// followed.
//
// If the copy fails half way through, the destination might be left
// partially written.
func Copy(srcFile, dstFile string) error {
	srcInfo, srcErr := os.Lstat(srcFile)
	if srcErr != nil {
		return srcErr
	}
	_, dstErr := os.Lstat(dstFile)
	if dstErr == nil {
		return fmt.Errorf("will not overwrite %q", dstFile)
	}
	if !os.IsNotExist(dstErr) {
		return dstErr
	}
	switch mode := srcInfo.Mode(); mode & os.ModeType {
	case os.ModeSymlink:
		return CopySymLink(srcFile, dstFile)
	case os.ModeDir:
		return CopyDir(srcFile, dstFile, mode)
	case 0:
		return CopyFile(srcFile, dstFile, mode)
	default:
		return fmt.Errorf("cannot copy file with mode %v", mode)
	}
}

func CopySymLink(srcFile, dstFile string) error {
	target, err := os.Readlink(srcFile)
	if err != nil {
		return err
	}
	return os.Symlink(target, dstFile)
}

func CopyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}
	defer dstFile.Close()
	// Make the actual permissions match the source permissions
	// even in the presence of umask.
	if err := os.Chmod(dstFile.Name(), mode.Perm()); err != nil {
		return err
	}
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("cannot copy %q to %q: %v", src, dst, err)
	}
	return nil
}

// CopyDir copy directory.
func CopyDir(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	if mode&0500 == 0 {
		// The source directory doesn't have write permission,
		// so give the new directory write permission anyway
		// so that we have permission to create its contents.
		// We'll make the permissions match at the end.
		mode |= 0500
	}
	if err := os.Mkdir(dst, mode.Perm()); err != nil {
		return err
	}
	for {
		names, err := srcFile.Readdirnames(100)
		for _, name := range names {
			if err := Copy(filepath.Join(src, name), filepath.Join(dst, name)); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading directory %q: %v", src, err)
		}
	}
	if err := os.Chmod(dst, mode.Perm()); err != nil {
		return err
	}
	return nil
}
