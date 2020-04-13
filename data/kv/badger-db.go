package kv

import (
	"github.com/angenalZZZ/gofunc/f"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"strconv"
	"time"
)

/**
 * Feature							键值功能设计	https://github.com/dgraph-io/badger
 * Design							数据结构设计	LSM tree with value log, it's not a B+ tree
 * High Read throughput				高读取吞吐量	Yes
 * High Write throughput			高写入吞吐量	Yes
 * Designed for SSDs				专为固态硬盘设计	Yes (with latest research 1)
 * Embeddable						可嵌入		Yes
 * Sorted KV access					排序KV访问	Yes
 * Transactions						事务管理		Yes, concurrent with SSI3, ACID: 原子性Atomicity、一致性Consistency、隔离性Isolation、持久性Durability
 * Snapshots						快照		Yes
 * TTL support						过期支持		Yes
 * 3D access (key-value-version)	3D访问（键值版本）Yes
 */
type BadgerDB struct {
	DB *badger.DB
}

// Open BadgerDB represents a badger db implementation,
// or no path, db save in memory.
func (db *BadgerDB) Open(path ...string) error {
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

	_db, err := badger.Open(opt)
	if err != nil {
		return err
	}

	db.DB = _db
	if opt.InMemory {
		return nil
	}

	go (func() {
		for db.DB.RunValueLogGC(0.5) == nil {
			// cleaning ...
		}
	})()
	return nil
}

// Size gets the size of the database (LSM + ValueLog) in bytes.
func (db *BadgerDB) Size() int64 {
	lsm, vLog := db.DB.Size()
	return lsm + vLog
}

// Incr - increment the key by the specified value.
func (db *BadgerDB) Incr(k string, by int64) (int64, error) {
	val, err := db.Get(k)
	if err != nil || val == "" {
		val = "0"
	}

	valFloat, _ := strconv.ParseInt(val, 10, 64)
	valFloat += by

	err = db.Set(k, strconv.FormatInt(valFloat, 10), -1)
	if err != nil {
		return 0, err
	}

	return valFloat, nil
}

// Set sets a key with the specified value and optional ttl.seconds
func (db *BadgerDB) Set(k, v string, ttl int) error {
	return db.SetBytes(f.Bytes(k), f.Bytes(v), ttl)
}

// SetBytes sets a key with the specified value and optional ttl.seconds
func (db *BadgerDB) SetBytes(k, v []byte, ttl int) error {
	return db.DB.Update(func(txn *badger.Txn) (err error) {
		if ttl <= 0 {
			err = txn.Set(k, v)
		} else {
			err = txn.SetEntry(&badger.Entry{
				Key:       k,
				Value:     v,
				UserMeta:  0,
				ExpiresAt: uint64(time.Now().Add(time.Duration(ttl) * time.Second).Unix()),
			})
		}
		return err
	})
}

// MSet sets multiple key-value pairs.
func (db *BadgerDB) MSet(data map[string]string) error {
	return db.DB.Update(func(txn *badger.Txn) (err error) {
		for k, v := range data {
			_ = txn.Set(f.Bytes(k), f.Bytes(v))
		}
		return nil
	})
}

// Get fetches the value of the specified k.
func (db *BadgerDB) Get(k string) (string, error) {
	var data string

	err := db.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(f.Bytes(k))
		if err != nil {
			return err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		data = f.String(val)
		return nil
	})
	return data, err
}

// MGet fetch multiple values of the specified keys.
func (db *BadgerDB) MGet(keys []string) (data []string) {
	data = make([]string, 0, len(keys))
	_ = db.DB.View(func(txn *badger.Txn) error {
		for _, key := range keys {
			item, err := txn.Get(f.Bytes(key))
			if err != nil {
				data = append(data, "")
				continue
			}
			val, err := item.ValueCopy(nil)
			if err != nil {
				data = append(data, "")
				continue
			}
			data = append(data, f.String(val))
		}
		return nil
	})
	return data
}

// TTL gets the time.seconds to live of the specified key's value.
func (db *BadgerDB) TTL(key string) int64 {
	var expires int64

	_ = db.DB.View(func(txn *badger.Txn) error {
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
		return -2
	}

	if expires == -1 {
		return -1
	}

	now := time.Now().Unix()

	if now >= expires {
		return -2
	}

	return (expires - now) / int64(time.Second)
}

// Del removes key(s) from the store.
func (db *BadgerDB) Del(keys []string) error {
	return db.DB.Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			_ = txn.Delete(f.Bytes(key))
		}
		return nil
	})
}

// Close ...
func (db *BadgerDB) Close() error {
	return db.DB.Close()
}

// Keys gets matched keys.
func (db *BadgerDB) Keys(prefix ...string) []string {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	var prefixBytes []byte
	var keys []string
	l := len(prefix)
	if l > 0 {
		prefixBytes = f.Bytes(prefix[0])
	}
	_ = db.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opts)
		defer it.Close()
		if l == 0 {
			for it.Rewind(); it.Valid(); it.Next() {
				keys = append(keys, f.String(it.Item().Key()))
			}
		} else {
			for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
				keys = append(keys, f.String(it.Item().Key()))
			}
		}
		return nil
	})
	return keys
}

// GC runs the garbage collector, not in memory.
func (db *BadgerDB) GC() error {
	var err error
	for {
		err = db.DB.RunValueLogGC(0.5)
		if err != nil {
			break
		}
	}
	return err
}
