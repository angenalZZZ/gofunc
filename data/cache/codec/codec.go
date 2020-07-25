package codec

import (
	"github.com/angenalZZZ/gofunc/data/cache/store"
)

// Stats allows to returns some statistics of codec usage
type Stats struct {
	Hits              int
	Miss              int
	SetSuccess        int
	SetError          int
	DeleteSuccess     int
	DeleteError       int
	InvalidateSuccess int
	InvalidateError   int
	ClearSuccess      int
	ClearError        int
}

// Codec represents an instance of a cache store
type Codec struct {
	store store.Interface
	stats *Stats
}

// New return a new codec instance
func New(store store.Interface) *Codec {
	return &Codec{
		store: store,
		stats: &Stats{},
	}
}

// Get allows to retrieve the value from a given key identifier
func (c *Codec) Get(key string) (interface{}, error) {
	val, err := c.store.Get(key)

	if err == nil {
		c.stats.Hits++
	} else {
		c.stats.Miss++
	}

	return val, err
}

// Set allows to set a value for a given key identifier and also allows to specify an expiration time
func (c *Codec) Set(key string, value interface{}, options *store.Options) error {
	err := c.store.Set(key, value, options)

	if err == nil {
		c.stats.SetSuccess++
	} else {
		c.stats.SetError++
	}

	return err
}

// TTL returns an expiration time
func (c *Codec) TTL(key string) int64 {
	return c.store.TTL(key)
}

// Delete allows to remove a value for a given key identifier
func (c *Codec) Delete(key string) error {
	err := c.store.Delete(key)

	if err == nil {
		c.stats.DeleteSuccess++
	} else {
		c.stats.DeleteError++
	}

	return err
}

// Invalidate invalidates some cach items from given options
func (c *Codec) Invalidate(options store.InvalidateOptions) error {
	err := c.store.Invalidate(options)

	if err == nil {
		c.stats.InvalidateSuccess++
	} else {
		c.stats.InvalidateError++
	}

	return err
}

// Clear resets all codec store data
func (c *Codec) Clear() error {
	err := c.store.Clear()

	if err == nil {
		c.stats.ClearSuccess++
	} else {
		c.stats.ClearError++
	}

	return err
}

// GetStore returns the store associated to this codec
func (c *Codec) GetStore() store.Interface {
	return c.store
}

// GetStats returns some statistics about the current codec
func (c *Codec) GetStats() *Stats {
	return c.stats
}
