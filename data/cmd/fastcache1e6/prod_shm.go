package main

import (
	"bytes"
	"fmt"
	"github.com/angenalZZZ/gofunc/data/shm"
	"github.com/angenalZZZ/gofunc/f"
	"io"
	"time"
)

func ProdSHM() {
	addr := fmt.Sprintf("%d", *flagPort)
	size := int32(1) << *flagCont
	if size < 2048 || size > 2097152 {
		size = int32(2097152)
	}

	const headSize = 4
	initHead := make([]byte, headSize)
	zeroHead := f.BytesRepeat('0', headSize)
	writeOff := int64(size / 2)

	m, err := shm.Create(addr, size)
	if err != nil {
		_ = fmt.Errorf("Io2SHM fail to create shared memory: %v\n", err)

		m, err = shm.Open(addr, size)
		if err != nil {
			_ = fmt.Errorf("Io2SHM fail to open shared memroy: %v\n", err)
			return
		}
	}
	fmt.Printf("Io2SHM server is listening on %s\n", addr)
	for {
		buf := make([]byte, headSize)
		n, err := m.ReadAt(buf, 0)
		if err != nil && err != io.EOF {
			_ = fmt.Errorf("Io2SHM failed to Read: %v\n", err)
			continue
		}
		if bytes.Equal(buf, initHead) || bytes.Equal(buf, zeroHead) {
			time.Sleep(time.Nanosecond)
			continue
		}
		fmt.Printf("%d< %s\n", 0, buf)

		buf = make([]byte, 80)
		n, _ = m.ReadAt(buf, 0)
		payload := make([]byte, n)
		copy(payload, buf)
		fmt.Printf("%d<< %s\n", 0, payload)

		buf1 := f.BytesRepeat('0', headSize)
		if n, err := m.WriteAt(buf1, 0); err != nil {
			_ = fmt.Errorf("Io2SHM failed to Write: %v\n", err)
		} else {
			fmt.Printf("%d> %s > len(%d)\n", 0, buf1, n)
		}
		buf2 := make([]byte, 80)
		if n, err := m.ReadAt(buf2, 0); err != nil && err != io.EOF {
			_ = fmt.Errorf("Io2SHM failed to Read: %v\n", err)
			continue
		} else {
			fmt.Printf("%d<< %s > len(%d)\n", 0, buf2, n)
		}

		result := defaultService.Handle(payload)
		if _, err := m.WriteAt(result, writeOff); err != nil {
			_ = fmt.Errorf("Io2SHM failed to Write: %v\n", err)
		} else {
			fmt.Printf("%d>> %s\n", writeOff, result)
		}
		return
	}
}
