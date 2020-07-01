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
		_ = fmt.Errorf("failed to serve: %v\n", err)
		return
	}

	fmt.Printf("server is listening on %s\n", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
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
				if err != nil || header.OpCode == ws.OpClose {
					return
				}

				payload := make([]byte, header.Length)
				_, err = io.ReadFull(conn, payload)
				if err != nil {
					continue
				}

				if header.Masked {
					ws.Cipher(payload, header.Mask, 0)
				}

				// Reset the Masked flag, server frames must not be masked as RFC6455 says.
				header.Masked = false

				if err = ws.WriteHeader(conn, header); err != nil || header.OpCode == ws.OpClose {
					return
				}
				result := defaultService.Handle(payload)
				if _, err := conn.Write(result); err != nil {
					continue
				}
			}
		}(conn)
	}
}
