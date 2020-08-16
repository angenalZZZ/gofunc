package codec

import (
	"github.com/angenalZZZ/gofunc/data/cache/store"
	"time"
)

// CodecInterface represents an instance of a cache codec
type CodecInterface interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, options *store.Options) error
	TTL(key string) (time.Duration, error)
	Delete(key string) error
	Invalidate(options store.InvalidateOptions) error
	Clear() error

	GetStore() store.Interface
	GetStats() *Stats
}
