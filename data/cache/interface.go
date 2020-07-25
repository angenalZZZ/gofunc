package cache

import (
	"github.com/angenalZZZ/gofunc/data/cache/codec"
	"github.com/angenalZZZ/gofunc/data/cache/store"
)

// Interface represents the interface for all caches (aggregates, metric, memory, redis, ...)
type Interface interface {
	Get(key string) (interface{}, error)
	Set(key string, object interface{}, options *store.Options) error
	TTL(key string) int64
	Delete(key string) error
	Invalidate(options store.InvalidateOptions) error
	Clear() error
	GetType() string
}

// StorageInterface represents the interface for caches that allows
// storage (for instance: memory, redis, ...)
type StorageInterface interface {
	Interface

	GetCodec() codec.CodecInterface
}
