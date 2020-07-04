package shm

import (
	"github.com/angenalZZZ/gofunc/f"
	"io"
	"os"
	"syscall"
)

type sharedMemory struct {
	h    syscall.Handle
	v    uintptr
	size int32
}

// create shared memory. return sharedMemory object.
func create(name string, size int32) (*sharedMemory, error) {
	key, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}

	h, err := syscall.CreateFileMapping(
		syscall.InvalidHandle, nil,
		syscall.PAGE_READWRITE, 0, uint32(size), key)
	if err != nil {
		return nil, os.NewSyscallError("CreateFileMapping", err)
	}

	v, err := syscall.MapViewOfFile(h, syscall.FILE_MAP_WRITE, 0, 0, 0)
	if err != nil {
		syscall.CloseHandle(h)
		return nil, os.NewSyscallError("MapViewOfFile", err)
	}

	return &sharedMemory{h, v, size}, nil
}

// open shared memory. return sharedMemory object.
func open(name string, size int32) (*sharedMemory, error) {
	return create(name, size)
}

func (o *sharedMemory) close() error {
	if o.v != uintptr(0) {
		syscall.UnmapViewOfFile(o.v)
		o.v = uintptr(0)
	}
	if o.h != syscall.InvalidHandle {
		syscall.CloseHandle(o.h)
		o.h = syscall.InvalidHandle
	}
	return nil
}

// read shared memory. return read size.
func (o *sharedMemory) readAt(p []byte, off int64) (n int, err error) {
	if off >= int64(o.size) {
		return 0, io.EOF
	}
	if max := int64(o.size) - off; int64(len(p)) > max {
		p = p[:max]
	}
	return f.BytesFromPtr(o.v, p, off, o.size), nil
}

// write shared memory. return write size.
func (o *sharedMemory) writeAt(p []byte, off int64) (n int, err error) {
	if off >= int64(o.size) {
		return 0, io.EOF
	}
	if max := int64(o.size) - off; int64(len(p)) > max {
		p = p[:max]
	}
	return f.BytesToPtr(p, o.v, off, o.size), nil
}
