package main

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/data/shm"
	"io"
)

func ProdSHM() {
	addr := fmt.Sprintf("%d", *flagPort)
	size := int32(1) << *flagCont
	if size < 4096 {
		size = int32(10485760)
	}

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
		buf := make([]byte, 4096)
		n, err := m.ReadAt(buf, 0)
		if err != nil && err != io.EOF {
			_ = fmt.Errorf("Io2SHM failed to Read: %v\n", err)
			continue
		}
		if n == 0 {
			continue
		}

		payload := buf[0:n]
		fmt.Println(payload)

		result := defaultService.Handle(payload)
		if _, err := m.Write(result); err != nil {
			_ = fmt.Errorf("Io2SHM failed to Write: %v\n", err)
		}
	}
}
