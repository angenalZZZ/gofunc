package main

import (
	"fmt"
	"github.com/gobwas/ws"
	"io"
	"net"
)

func ProdWS() {
	addr := fmt.Sprintf("127.0.0.1:%d", *flagPort)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		_ = fmt.Errorf("Io2WS failed to serve: %v\n", err)
		return
	}
	fmt.Printf("Io2WS server is listening on %s\n", addr)
	for {
		conn, err := l.Accept()
		if err != nil {
			_ = fmt.Errorf("Io2WS failed to serve: %v\n", err)
			continue
		}
		_, err = ws.Upgrade(conn)
		if err != nil {
			continue
		}

		go func(conn net.Conn) {
			defer conn.Close()

			for {
				header, err := ws.ReadHeader(conn)
				if err != nil {
					_ = fmt.Errorf("Io2WS failed to ReadHeader: %v\n", err)
					return
				}
				if header.OpCode == ws.OpClose {
					_ = fmt.Errorf("Io2WS client to Close: %v\n", err)
					return
				}

				payload := make([]byte, header.Length)
				_, err = io.ReadFull(conn, payload)
				if err != nil {
					_ = fmt.Errorf("Io2WS failed to serve: %v\n", err)
					continue
				}

				if header.Masked {
					ws.Cipher(payload, header.Mask, 0)
				}

				// Reset the Masked flag, server frames must not be masked as RFC6455 says.
				header.Masked = false
				if err = ws.WriteHeader(conn, header); err != nil {
					_ = fmt.Errorf("Io2WS failed to WriteHeader: %v\n", err)
					return
				}

				result := defaultService.Handle(payload)
				if _, err := conn.Write(result); err != nil {
					_ = fmt.Errorf("Io2WS failed to Write: %v\n", err)
				}
			}
		}(conn)
	}
}
