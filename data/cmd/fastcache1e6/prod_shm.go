// install https://github.com/angenalZZZ/doc-zip/raw/master/nanomsg-mingw64.7z
// go get github.com/op/go-nanomsg
// INPROC - transport within a process (between threads, modules etc.)
// IPC - transport between processes on a single machine
// TCP - network transport via TCP
// WS - websockets over TCP

package main

import (
	"bytes"
	"fmt"
	"github.com/angenalZZZ/gofunc/data/shm"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/op/go-nanomsg"
	"io"
	"strings"
	"syscall"
	"time"
)

func ProdSHM() {
	const headSize = 4
	addr := *flagAddr
	if addr == "" {
		_ = fmt.Errorf("Io2SHM fail to check the server address -a has prefix ipc://\n")
		return
	}

	if strings.HasPrefix(addr, "ipc://") {
		// github.com/op/go-nanomsg
		var (
			err error
			rep *nanomsg.Socket
		)

		if rep, err = nanomsg.NewSocket(nanomsg.AF_SP, nanomsg.REP); err != nil {
			_ = fmt.Errorf("Io2SHM fail to create shared memory: %v\n", err)
			return
		}
		if _, err = rep.Bind(addr); err != nil {
			_ = fmt.Errorf("Io2SHM fail to create shared memory: %v\n", err)
			return
		}

		defer func() { _ = rep.Close() }()
		if err = rep.SetRecvTimeout(time.Millisecond); err != nil {
			return
		}

		for {
			if payload, err := rep.Recv(0); err != nil {
				if err == syscall.ETIMEDOUT || err == io.EOF {
					continue
				}
				_ = fmt.Errorf("Io2SHM failed to Read: %v\n", err)
			} else if len(payload) > headSize {
				fmt.Printf("Io2SHM Recv %q\n", string(payload))
				result := defaultService.Handle(payload)
				if _, err = rep.Send(result, 0); err != nil {
					_ = fmt.Errorf("Io2SHM failed to Write: %v\n", err)
				} else {
					fmt.Printf("Io2SHM Send %q\n", string(result))
				}
			}
		}
	} else {
		// github.com/angenalZZZ/gofunc/data/shm
		size := int32(1) << *flagCont
		if size < 2048 || size > 2097152 {
			size = int32(2097152)
		}

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
}
