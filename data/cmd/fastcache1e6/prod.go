package main

import (
	"github.com/angenalZZZ/gofunc/g"
)

func Prod() {
	// Init cache instance
	defaultService = &cacheServiceImpl{}

	// IO transmission mode SHM(SharedMemory)/gRPC/TCP/WS(WebSocket)/NatS
	switch *flagSvc {
	case g.Io2SHM:
		ProdSHM()
	case g.Io2gRPC:
		ProdGRPC()
	case g.Io2TCP:
		ProdTCP()
	case g.Io2WS:
		ProdWS()
	case g.Io2NatS:
		ProdNatS()
	}
}
