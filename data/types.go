package data

// Map is a shortcut for map[string]interface{}
type Map map[string]interface{}

// IO transmission mode SHM(SharedMemory)/gRPC/TCP/WS(WebSocket)
const (
	Io2SHM int = iota
	Io2gRPC
	Io2TCP
	Io2WS
)
