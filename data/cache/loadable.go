package cache

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/data/cache/store"
)

const (
	// LoadableType represents the loadable cache type as a string value
	LoadableType = "loadable"
)

type loadableKeyValue struct {
	key   string
	value interface{}
}

type loadFunction func(key interface{}) (interface{}, error)

// LoadableCache represents a cache that uses a function to load data
type LoadableCache struct {
	loadFunc   loadFunction
	cache      Interface
	setChannel chan *loadableKeyValue
}

// NewLoadable create a new cache that uses a function to load data
func NewLoadable(loadFunc loadFunction, cache Interface) *LoadableCache {
	loadable := &LoadableCache{
		loadFunc:   loadFunc,
		cache:      cache,
		setChannel: make(chan *loadableKeyValue, 10000),
	}

	go loadable.setter()

	return loadable
}

func (c *LoadableCache) setter() {
	for item := range c.setChannel {
		_ = c.Set(item.key, item.value, nil)
	}
}

// Get returns the object stored in cache if it exists
func (c *LoadableCache) Get(key string) (object interface{}, err error) {
	object, err = c.cache.Get(key)
	if err == nil {
		return
	}

	// Unable to find in cache, try to load it from load function
	object, err = c.loadFunc(key)
	if err != nil {
		_ = fmt.Errorf("An error has occurred while trying to load item from load function: %v\n", err)
		return
	}

	// Then, put it back in cache
	c.setChannel <- &loadableKeyValue{key, object}
	return
}

// Set sets a value in available caches
func (c *LoadableCache) Set(key string, object interface{}, options *store.Options) error {
	return c.cache.Set(key, object, options)
}

// TTL returns an expiration time
func (c *LoadableCache) TTL(key string) int64 {
	return c.cache.TTL(key)
}

// Delete removes a value from cache
func (c *LoadableCache) Delete(key string) error {
	return c.cache.Delete(key)
}

// Invalidate invalidates cache item from given options
func (c *LoadableCache) Invalidate(options store.InvalidateOptions) error {
	return c.cache.Invalidate(options)
}

// Clear resets all cache data
func (c *LoadableCache) Clear() error {
	return c.cache.Clear()
}

// GetType returns the cache type
func (c *LoadableCache) GetType() string {
	return LoadableType
}
