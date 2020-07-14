package cache

import (
	"github.com/angenalZZZ/gofunc/data/codec"
	"github.com/angenalZZZ/gofunc/data/store"
	"github.com/angenalZZZ/gofunc/f"
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
func (c *Cache) Get(key interface{}) (interface{}, error) {
	cacheKey := c.getMD5Key(key)
	return c.codec.Get(cacheKey)
}

// Set populates the cache item using the given key
func (c *Cache) Set(key, object interface{}, options *store.Options) error {
	cacheKey := c.getMD5Key(key)
	return c.codec.Set(cacheKey, object, options)
}

// Delete removes the cache item using the given key
func (c *Cache) Delete(key interface{}) error {
	cacheKey := c.getMD5Key(key)
	return c.codec.Delete(cacheKey)
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

// getMD5Key returns the cache key for the given key object by computing a checksum of key struct
func (c *Cache) getMD5Key(key interface{}) string {
	return f.CryptoMD5Key(key)
}
