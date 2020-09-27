package main

import (
	. "github.com/angenalZZZ/gofunc/g"
)

func Prod() {
	// Init cache instance
	defaultService = &cacheServiceImpl{}

	// IO transmission mode SHM(SharedMemory)/gRPC/TCP/WS(WebSocket)/NatS
	switch *flagSvc {
	case Io2SHM:
		ProdSHM()
	case Io2gRPC:
		ProdGRPC()
	case Io2TCP:
		ProdTCP()
	case Io2WS:
		ProdWS()
	case Io2NatS:
		ProdNatS()
	}
}
