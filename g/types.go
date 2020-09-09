package g

// Strings is a string array
type Strings []string

// IO transmission mode SHM(SharedMemory)/gRPC/TCP/WS(WebSocket)/NatS
const (
	Io2SHM int = iota
	Io2gRPC
	Io2TCP
	Io2WS
	Io2NatS
)
