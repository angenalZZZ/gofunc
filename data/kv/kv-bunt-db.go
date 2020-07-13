package kv

import (
	"github.com/angenalZZZ/gofunc/f"
	"github.com/tidwall/buntdb"
	"strconv"
	"strings"
	"time"
)

// BuntDB is a low-level, in-memory, key/value store.
// It persists to disk, is ACID compliant, and uses locking for multiple readers and a single writer.
// It supports custom indexes and geo-spatial data.
// It's ideal for projects that need a dependable database and favor speed over data size.
type BuntDB struct {
	DB     *buntdb.DB
	locker *f.Locker
}

// Open Opens the specified path, default in memory.
func (db *BuntDB) Open(path ...string) error {
	filename := ":memory:"
	if len(path) > 0 {
		filename = path[0]
	} else {
		//filename, _ = ioutil.TempDir(os.TempDir(), "")
	}
	var err error
	db.DB, err = buntdb.Open(filename)
	if err != nil {
		return err
	}
	db.locker = f.NewLocker()
	return nil
}

// Size gets the size of the database in bytes.
func (db *BuntDB) Size() int64 {
	return 0
}

// Incr increment the key by the specified value.
func (db *BuntDB) Incr(k string, by int64) (int64, error) {
	db.locker.Lock(k)
	defer db.locker.Unlock(k)

	val, err := db.get(k)
	if err != nil || val == "" {
		val = "0"
	}

	valFloat, _ := strconv.ParseInt(val, 10, 64)
	valFloat += by

	err = db.set(k, strconv.FormatInt(valFloat, 10), -1)
	if err != nil {
		return 0, err
	}

	return valFloat, nil
}

// Set sets a key with the specified value and optional ttl.seconds
func (db *BuntDB) Set(k, v string, ttl int) error {
	return db.set(k, v, ttl)
}

// SetBytes sets a key with the specified value and optional ttl.seconds
func (db *BuntDB) SetBytes(k, v []byte, ttl int) error {
	return db.set(f.String(k), f.String(v), ttl)
}

// MSet sets multiple key-value pairs.
func (db *BuntDB) MSet(data map[string]string) (err error) {
	err = db.DB.Update(func(tx *buntdb.Tx) (err1 error) {
		for k, v := range data {
			_, _, err1 = tx.Set(k, v, nil)
		}
		return
	})
	return
}

// Get fetches the value of the specified k.
func (db *BuntDB) Get(k string) (string, error) {
	return db.get(k)
}

// MGet fetch multiple values of the specified keys.
func (db *BuntDB) MGet(keys []string) (data []string) {
	data = make([]string, 0, len(keys))
	for _, key := range keys {
		val, err := db.get(key)
		if err != nil {
			data = append(data, "")
			continue
		}
		data = append(data, val)
	}
	return data
}

// TTL gets the time.seconds to live of the specified key's value.
func (db *BuntDB) TTL(key string) int64 {
	ts := time.Nanosecond
	_ = db.DB.View(func(tx *buntdb.Tx) (err1 error) {
		ts, err1 = tx.TTL(key)
		return
	})
	return int64(ts.Seconds())
}

// Del removes key(s) from the store.
func (db *BuntDB) Del(keys []string) (err error) {
	err = db.DB.Update(func(tx *buntdb.Tx) (err1 error) {
		for _, k := range keys {
			_, err1 = tx.Delete(k)
		}
		return
	})
	return
}

// Close ...
func (db *BuntDB) Close() error {
	return db.DB.Close()
}

// Keys gets matched keys.
func (db *BuntDB) Keys(prefix ...string) []string {
	var prefix1 string
	var keys []string
	l := len(prefix)
	if l > 0 {
		prefix1 = prefix[0]
		l = len(prefix1)
	}
	_ = db.DB.View(func(tx *buntdb.Tx) (err1 error) {
		err1 = tx.AscendKeys("", func(key, _ string) bool {
			if l == 0 || strings.HasPrefix(key, prefix1) {
				keys = append(keys, key)
			}
			return true
		})
		return
	})
	return keys
}

// GC runs the garbage collector.
func (db *BuntDB) GC() error {
	return db.DB.Shrink()
}

func (db *BuntDB) get(k string) (data string, err error) {
	err = db.DB.View(func(tx *buntdb.Tx) (err1 error) {
		data, err1 = tx.Get(k)
		return
	})
	return
}

func (db *BuntDB) set(k, v string, ttl int) (err error) {
	err = db.DB.Update(func(tx *buntdb.Tx) (err1 error) {
		_, _, err1 = tx.Set(k, v, &buntdb.SetOptions{Expires: true, TTL: time.Duration(ttl) * time.Second})
		return
	})
	return
}
