package cache

import (
	"github.com/angenalZZZ/gofunc/data/cache/metrics"
	"github.com/angenalZZZ/gofunc/data/cache/store"
)

const (
	// MetricType represents the metric cache type as a string value
	MetricType = "metric"
)

// MetricCache is the struct that specifies metrics available for different caches
type MetricCache struct {
	metrics metrics.MetricsInterface
	cache   Interface
}

// NewMetric creates a new cache with metrics and a given cache storage
func NewMetric(metrics metrics.MetricsInterface, cache Interface) *MetricCache {
	return &MetricCache{
		metrics: metrics,
		cache:   cache,
	}
}

// Get obtains a value from cache and also records metrics
func (c *MetricCache) Get(key string) (interface{}, error) {
	result, err := c.cache.Get(key)

	c.updateMetrics(c.cache)

	return result, err
}

// Set sets a value from the cache
func (c *MetricCache) Set(key string, object interface{}, options *store.Options) error {
	return c.cache.Set(key, object, options)
}

// TTL returns an expiration time
func (c *MetricCache) TTL(key string) int64 {
	return c.cache.TTL(key)
}

// Delete removes a value from the cache
func (c *MetricCache) Delete(key string) error {
	return c.cache.Delete(key)
}

// Invalidate invalidates cache item from given options
func (c *MetricCache) Invalidate(options store.InvalidateOptions) error {
	return c.cache.Invalidate(options)
}

// Clear resets all cache data
func (c *MetricCache) Clear() error {
	return c.cache.Clear()
}

// Get obtains a value from cache and also records metrics
func (c *MetricCache) updateMetrics(cache Interface) {
	switch current := cache.(type) {
	case *ChainCache:
		for _, cache := range current.GetCaches() {
			c.updateMetrics(cache)
		}

	case StorageInterface:
		c.metrics.RecordFromCodec(current.GetCodec())
	}
}

// GetType returns the cache type
func (c *MetricCache) GetType() string {
	return MetricType
}
