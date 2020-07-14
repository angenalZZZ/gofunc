package cache

import (
	"github.com/angenalZZZ/gofunc/data/codec"
	"github.com/angenalZZZ/gofunc/data/store"
)

// Interface represents the interface for all caches (aggregates, metric, memory, redis, ...)
type Interface interface {
	Get(key interface{}) (interface{}, error)
	Set(key, object interface{}, options *store.Options) error
	Delete(key interface{}) error
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
