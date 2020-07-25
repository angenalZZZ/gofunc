package kv

// KV key value database interface
// Feature github.com/angenalZZZ/gofunc/data/kv/...
type KV interface {
	Open(...string) error
	Size() int64
	Incr(string, int64) (int64, error)
	Set(string, string, int) error
	SetBytes([]byte, []byte, int) error
	MSet(map[string]string) error
	Get(string) (string, error)
	GetBytes([]byte) ([]byte, error)
	MGet([]string) []string
	TTL(string) int64
	Del([]string) error
	Close() error
	Keys(...string) []string
	GC() error
}
