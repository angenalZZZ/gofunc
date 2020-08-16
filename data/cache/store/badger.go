package store

import (
	"errors"
	"fmt"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"strings"
	"time"
)

const (
	// BadgerType represents the storage type as a string value
	BadgerType = "badger"
	// BadgerTagPattern represents the tag pattern to be used as a key in specified storage
	BadgerTagPattern = "gocache_tag_%s"
)

// BadgerStore is a store for Redis
type BadgerStore struct {
	// BadgerClientInterface represents a dgraph-io/badger client
	client  *badger.DB
	options *Options
}

// NewBadger creates a new store to Badger instance(s)
func NewBadger(option *Options, path ...string) *BadgerStore {
	if option == nil {
		option = &Options{}
	}

	var opt badger.Options
	if len(path) == 1 {
		opt = badger.DefaultOptions(path[0])
		opt.Truncate = true
		opt.SyncWrites = false
		opt.TableLoadingMode = options.MemoryMap
		opt.ValueLogLoadingMode = options.FileIO
		//opt.ValueThreshold = 1 << 20 // 阈值 1 MB
		opt.ValueThreshold = 1 // 阈值 默认 32
		opt.NumMemtables = 2
		opt.NumLevelZeroTables = 2
		opt.MaxTableSize = 16 << 20
	} else {
		opt = badger.DefaultOptions("").WithInMemory(true)
	}

	client, err := badger.Open(opt)
	if err != nil {
		return nil
	}

	if opt.InMemory == false {
		go (func() {
			for client.RunValueLogGC(0.5) == nil {
				// cleaning ...
			}
		})()
	}

	return &BadgerStore{
		client:  client,
		options: option,
	}
}

// Get returns data stored from a given key
func (s *BadgerStore) Get(key string) (interface{}, error) {
	var data []byte
	err := s.client.View(func(txn *badger.Txn) error {
		item, err := txn.Get(f.Bytes(key))
		if err != nil {
			return err
		}

		data, err = item.ValueCopy(nil)
		return err
	})
	return data, err
}

// TTL returns a expiration time
func (s *BadgerStore) TTL(key string) (time.Duration, error) {
	var expires int64

	_ = s.client.View(func(txn *badger.Txn) error {
		item, err := txn.Get(f.Bytes(key))
		if err != nil {
			expires = -2
			return nil
		}

		exp := item.ExpiresAt()
		if exp == 0 {
			expires = -1
			return nil
		}

		expires = int64(exp)
		return nil
	})

	if expires == -2 {
		return -2, errors.New("unable to retrieve data from badger")
	}

	if expires == -1 {
		return -1, errors.New("unable to retrieve data from badger")
	}

	now := time.Now().Unix()

	if now >= expires {
		return -2, errors.New("unable to retrieve data from badger")
	}

	return time.Second * time.Duration((expires-now)/int64(time.Second)), nil
}

// Set defines data in Redis for given key identifier
func (s *BadgerStore) Set(key string, value interface{}, options *Options) error {
	if options == nil {
		options = s.options
	}

	err := s.client.Update(func(txn *badger.Txn) (err error) {
		if options.Expiration <= 0 {
			err = txn.Set(f.Bytes(key), value.([]byte))
		} else {
			err = txn.SetEntry(&badger.Entry{
				Key:       f.Bytes(key),
				Value:     value.([]byte),
				ExpiresAt: uint64(time.Now().Add(options.Expiration).Unix()),
			})
		}
		return err
	})

	if err != nil {
		return err
	}

	if tags := options.TagsValue(); len(tags) > 0 {
		s.setTags(key, tags)
	}

	return nil
}

func (s *BadgerStore) setTags(key string, tags []string) {
	for _, tag := range tags {
		var tagKey = fmt.Sprintf(BadgerTagPattern, tag)
		var cacheKeys []string

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

		_ = s.Set(tagKey, []byte(strings.Join(cacheKeys, ",")), &Options{
			Expiration: 720 * time.Hour,
		})
	}
}

// Delete removes data from Redis for given key identifier
func (s *BadgerStore) Delete(key string) error {
	return s.client.Update(func(txn *badger.Txn) error {
		return txn.Delete(f.Bytes(key))
	})
}

// Invalidate invalidates some cache data in Redis for given options
func (s *BadgerStore) Invalidate(options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			var tagKey = fmt.Sprintf(BadgerTagPattern, tag)
			result, err := s.Get(tagKey)
			if err != nil {
				return nil
			}

			var cacheKeys []string
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}

			for _, cacheKey := range cacheKeys {
				_ = s.Delete(cacheKey)
			}
		}
	}

	return nil
}

// Clear resets all data in the store
func (s *BadgerStore) Clear() error {
	return s.client.DropAll()
}

// GetType returns the store type
func (s *BadgerStore) GetType() string {
	return BadgerType
}
