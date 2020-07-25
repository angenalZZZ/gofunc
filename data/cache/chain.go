package cache

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/data/cache/store"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/panjf2000/ants/v2"
)

const (
	// ChainType represents the chain cache type as a string value
	ChainType = "chain"
)

// chainKeyValue transport in the channel
type chainKeyValue struct {
	idx     int
	key     interface{}
	value   interface{}
	options *store.Options
}

// ChainCache represents the configuration needed by a cache aggregator
type ChainCache struct {
	caches []StorageInterface
	pool   *ants.PoolWithFunc
}

// NewChain create a new cache aggregator
func NewChain(caches ...StorageInterface) *ChainCache {
	chain := &ChainCache{caches: caches}

	chain.pool, _ = ants.NewPoolWithFunc(100000, func(payload interface{}) {
		if payload == nil {
			return
		}
		if set, ok := payload.(*chainKeyValue); ok {
			_ = chain.caches[set.idx].Set(set.key, set.value, set.options)
		}
	}, ants.WithOptions(ants.Options{
		ExpiryDuration:   ants.DefaultCleanIntervalTime,
		PreAlloc:         true,
		Nonblocking:      true,
		MaxBlockingTasks: 0,
		PanicHandler: func(err interface{}) {
			_ = fmt.Errorf(" GoHttpHandle/worker: %s\n %v", f.Now().LocalTimeString(), err)
		},
	}))
	return chain
}

// Get returns the value stored in cache if it exists
func (c *ChainCache) Get(key string) (value interface{}, err error) {
	for i, cache := range c.caches {
		value, err = cache.Get(key)
		if err == nil {
			// Set the value back until this cache layer
			if err = c.pool.Invoke(&chainKeyValue{
				idx:     i,
				key:     key,
				value:   value,
				options: nil, // TODO: get store Options
			}); err != nil {
				_ = fmt.Errorf("unable to set item into cache with store '%s': %v", cache.GetCodec().GetStore().GetType(), err)
			}
			return value, nil
		}

		_ = fmt.Errorf("Unable to retrieve item from cache with store '%s': %v\n", cache.GetCodec().GetStore().GetType(), err)
	}
	return
}

// Set sets a value in available caches
func (c *ChainCache) Set(key string, value interface{}, options *store.Options) error {
	for i, cache := range c.caches {
		if i == 0 && options.Async == false {
			err := cache.Set(key, value, options)
			if err != nil {
				storeType := cache.GetCodec().GetStore().GetType()
				return fmt.Errorf("unable to set item into cache with store '%s': %v", storeType, err)
			}
		} else {
			if err := c.pool.Invoke(&chainKeyValue{
				idx:     i,
				key:     key,
				value:   value,
				options: options,
			}); err != nil {
				storeType := cache.GetCodec().GetStore().GetType()
				return fmt.Errorf("unable to set item into cache with store '%s': %v", storeType, err)
			}
		}
	}
	return nil
}

// Delete removes a value from all available caches
func (c *ChainCache) Delete(key string) error {
	for _, cache := range c.caches {
		_ = cache.Delete(key)
	}
	return nil
}

// TTL returns an expiration time
func (c *ChainCache) TTL(key string) int64 {
	return c.caches[len(c.caches)-1].TTL(key)
}

// Invalidate invalidates cache item from given options
func (c *ChainCache) Invalidate(options store.InvalidateOptions) error {
	for _, cache := range c.caches {
		_ = cache.Invalidate(options)
	}
	return nil
}

// Clear resets all cache data
func (c *ChainCache) Clear() error {
	c.pool.Release()
	for _, cache := range c.caches {
		_ = cache.Clear()
	}
	c.pool.Reboot()
	return nil
}

// GetCaches returns all chain caches
func (c *ChainCache) GetCaches() []StorageInterface {
	return c.caches
}

// GetType returns the cache type
func (c *ChainCache) GetType() string {
	return ChainType
}
