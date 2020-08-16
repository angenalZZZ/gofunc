package store

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

const (
	// MemcacheType represents the storage type as a string value
	MemcacheType = "memcache"
	// MemcacheTagPattern represents the tag pattern to be used as a key in specified storage
	MemcacheTagPattern = "gocache_tag_%s"
)

// MemcacheStore is a store for Memcache
type MemcacheStore struct {
	// MemcacheClientInterface represents a bradfitz/gomemcache client
	client  *memcache.Client
	options *Options
}

// NewMemcache creates a new store to Memcache instance(s)
func NewMemcache(client *memcache.Client, options *Options) *MemcacheStore {
	if options == nil {
		options = &Options{}
	}

	return &MemcacheStore{
		client:  client,
		options: options,
	}
}

// Get returns data stored from a given key
func (s *MemcacheStore) Get(key string) (interface{}, error) {
	item, err := s.client.Get(key)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("unable to retrieve data from memcache")
	}

	return item.Value, err
}

// TTL returns a expiration time
func (s *MemcacheStore) TTL(key string) (time.Duration, error) {
	item, err := s.client.Get(key)
	if err != nil {
		return 0, err
	}
	if item == nil {
		return 0, errors.New("unable to retrieve data from memcache")
	}

	return time.Second * time.Duration(item.Expiration), err
}

// Set defines data in Memcache for given key identifier
func (s *MemcacheStore) Set(key string, value interface{}, options *Options) error {
	if options == nil {
		options = s.options
	}

	item := &memcache.Item{
		Key:        key,
		Value:      value.([]byte),
		Expiration: int32(options.ExpirationValue().Seconds()),
	}

	err := s.client.Set(item)
	if err != nil {
		return err
	}

	if tags := options.TagsValue(); len(tags) > 0 {
		s.setTags(key, tags)
	}

	return nil
}

func (s *MemcacheStore) setTags(key string, tags []string) {
	for _, tag := range tags {
		var tagKey = fmt.Sprintf(MemcacheTagPattern, tag)
		var cacheKeys = []string{}

		if result, err := s.Get(tagKey); err == nil {
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}
		}

		var alreadyInserted = false
		for _, cacheKey := range cacheKeys {
			if cacheKey == key {
				alreadyInserted = true
				break
			}
		}

		if !alreadyInserted {
			cacheKeys = append(cacheKeys, key)
		}

		s.Set(tagKey, []byte(strings.Join(cacheKeys, ",")), &Options{
			Expiration: 720 * time.Hour,
		})
	}
}

// Delete removes data from Memcache for given key identifier
func (s *MemcacheStore) Delete(key string) error {
	return s.client.Delete(key)
}

// Invalidate invalidates some cache data in Redis for given options
func (s *MemcacheStore) Invalidate(options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			var tagKey = fmt.Sprintf(MemcacheTagPattern, tag)
			result, err := s.Get(tagKey)
			if err != nil {
				return nil
			}

			var cacheKeys = []string{}
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}

			for _, cacheKey := range cacheKeys {
				s.Delete(cacheKey)
			}
		}
	}

	return nil
}

// Clear resets all data in the store
func (s *MemcacheStore) Clear() error {
	return s.client.FlushAll()
}

// GetType returns the store type
func (s *MemcacheStore) GetType() string {
	return MemcacheType
}
