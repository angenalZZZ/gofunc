package shm

import (
	"fmt"
	"io"
)

// SharedMemory is shared memory struct
type SharedMemory struct {
	m   *sharedMemory
	pos int64
}

// Create is create shared memory
func Create(name string, size int32) (*SharedMemory, error) {
	m, err := create(name, size)
	if err != nil {
		return nil, err
	}
	return &SharedMemory{m, 0}, nil
}

// Open is open exist shared memory
func Open(name string, size int32) (*SharedMemory, error) {
	m, err := open(name, size)
	if err != nil {
		return nil, err
	}
	return &SharedMemory{m, 0}, nil
}

// Close is close & discard shared memory
func (o *SharedMemory) Close() (err error) {
	if o.m != nil {
		err = o.m.close()
		if err == nil {
			o.m = nil
		}
	}
	return err
}

// Read is read shared memory (current position)
func (o *SharedMemory) Read(p []byte) (n int, err error) {
	n, err = o.ReadAt(p, o.pos)
	if err != nil {
		return 0, err
	}
	o.pos += int64(n)
	return n, nil
}

// ReadAt is read shared memory (offset)
func (o *SharedMemory) ReadAt(p []byte, off int64) (n int, err error) {
	return o.m.readAt(p, off)
}

// Seek is move read/write position at shared memory
func (o *SharedMemory) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		offset += int64(0)
	case io.SeekCurrent:
		offset += o.pos
	case io.SeekEnd:
		offset += int64(o.m.size)
	}
	if offset < 0 || offset >= int64(o.m.size) {
		return 0, fmt.Errorf("invalid offset")
	}
	o.pos = offset
	return offset, nil
}

// Write is write shared memory (current position)
func (o *SharedMemory) Write(p []byte) (n int, err error) {
	n, err = o.WriteAt(p, o.pos)
	if err != nil {
		return 0, err
	}
	o.pos += int64(n)
	return n, nil
}

// WriteAt is write shared memory (offset)
func (o *SharedMemory) WriteAt(p []byte, off int64) (n int, err error) {
	return o.m.writeAt(p, off)
}
