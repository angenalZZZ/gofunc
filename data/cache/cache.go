package cache

import (
	"github.com/angenalZZZ/gofunc/data/cache/codec"
	"github.com/angenalZZZ/gofunc/data/cache/store"
)

const (
	// DefaultType represents the cache type as a string value
	DefaultType = "cache"
)

// Cache represents the configuration needed by a cache
type Cache struct {
	codec codec.CodecInterface
}

// New create a new cache entry
func New(store store.Interface) *Cache {
	return &Cache{
		codec: codec.New(store),
	}
}

// Get returns the object stored in cache if it exists
func (c *Cache) Get(key string) (interface{}, error) {
	return c.codec.Get(key)
}

// Set populates the cache item using the given key
func (c *Cache) Set(key string, object interface{}, options *store.Options) error {
	return c.codec.Set(key, object, options)
}

// TTL returns an expiration time
func (c *Cache) TTL(key string) int64 {
	return c.codec.TTL(key)
}

// Delete removes the cache item using the given key
func (c *Cache) Delete(key string) error {
	return c.codec.Delete(key)
}

// Invalidate invalidates cache item from given options
func (c *Cache) Invalidate(options store.InvalidateOptions) error {
	return c.codec.Invalidate(options)
}

// Clear resets all cache data
func (c *Cache) Clear() error {
	return c.codec.Clear()
}

// GetCodec returns the current codec
func (c *Cache) GetCodec() codec.CodecInterface {
	return c.codec
}

// GetType returns the cache type
func (c *Cache) GetType() string {
	return DefaultType
}
