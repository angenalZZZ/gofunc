package main

import "github.com/angenalZZZ/gofunc/data"

func Prod() {
	// Init cache instance
	defaultService = &cacheServiceImpl{}

	// IO transmission mode SHM(SharedMemory)/gRPC/TCP/WS(WebSocket)
	switch *flagSvc {
	case data.Io2SHM:
		ProdSHM()
	case data.Io2gRPC:
		ProdGRPC()
	case data.Io2TCP:
		ProdTCP()
	case data.Io2WS:
		ProdWS()
	}
}
