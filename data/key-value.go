package data

// key value database interface
// Feature github.com/angenalZZZ/gofunc/data/kv/...
type KvDB interface {
	Size() int64
	GC() error
	Incr(string, int64) (int64, error)
	Set(string, string, int) error
	SetBytes([]byte, []byte, int) error
	MSet(map[string]string) error
	Get(string) (string, error)
	MGet([]string) []string
	TTL(string) int64
	Del([]string) error
}
