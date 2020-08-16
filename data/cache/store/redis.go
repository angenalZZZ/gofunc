package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
)

const (
	// RedisType represents the storage type as a string value
	RedisType = "redis"
	// RedisTagPattern represents the tag pattern to be used as a key in specified storage
	RedisTagPattern = "gocache_tag_%s"
)

// RedisStore is a store for Redis
type RedisStore struct {
	// RedisClientInterface represents a go-redis/redis client
	client  *redis.Client
	options *Options
}

// NewRedis creates a new store to Redis instance(s)
func NewRedis(client *redis.Client, options *Options) *RedisStore {
	if options == nil {
		options = &Options{}
	}

	return &RedisStore{
		client:  client,
		options: options,
	}
}

// Get returns data stored from a given key
func (s *RedisStore) Get(key string) (interface{}, error) {
	return s.client.Get(key).Result()
}

// TTL returns a expiration time
func (s *RedisStore) TTL(key string) (time.Duration, error) {
	return s.client.TTL(key).Result()
}

// Set defines data in Redis for given key identifier
func (s *RedisStore) Set(key string, value interface{}, options *Options) error {
	if options == nil {
		options = s.options
	}

	err := s.client.Set(key, value, options.ExpirationValue()).Err()
	if err != nil {
		return err
	}

	if tags := options.TagsValue(); len(tags) > 0 {
		s.setTags(key, tags)
	}

	return nil
}

func (s *RedisStore) setTags(key string, tags []string) {
	for _, tag := range tags {
		var tagKey = fmt.Sprintf(RedisTagPattern, tag)
		var cacheKeys = s.getCacheKeysForTag(tagKey)

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

		_ = s.Set(tagKey, strings.Join(cacheKeys, ","), &Options{
			Expiration: 720 * time.Hour,
		})
	}
}

func (s *RedisStore) getCacheKeysForTag(tagKey string) []string {
	var cacheKeys []string
	if result, err := s.Get(tagKey); err == nil && result != "" {
		if str, ok := result.(string); ok {
			cacheKeys = strings.Split(str, ",")
		}
	}
	return cacheKeys
}

// Delete removes data from Redis for given key identifier
func (s *RedisStore) Delete(key string) error {
	_, err := s.client.Del(key).Result()
	return err
}

// Invalidate invalidates some cache data in Redis for given options
func (s *RedisStore) Invalidate(options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			var tagKey = fmt.Sprintf(RedisTagPattern, tag)
			var cacheKeys = s.getCacheKeysForTag(tagKey)

			for _, cacheKey := range cacheKeys {
				_ = s.Delete(cacheKey)
			}

			_ = s.Delete(tagKey)
		}
	}

	return nil
}

// GetType returns the store type
func (s *RedisStore) GetType() string {
	return RedisType
}

// Clear resets all data in the store
func (s *RedisStore) Clear() error {
	if err := s.client.FlushAll().Err(); err != nil {
		return err
	}

	return nil
}
