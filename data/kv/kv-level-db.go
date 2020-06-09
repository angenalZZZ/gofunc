package kv

import (
	"errors"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

// LevelDB represents a leveldb db implementation.
type LevelDB struct {
	DB     *leveldb.DB
	locker *f.Locker
}

// Open Opens the specified path.
func (db *LevelDB) Open(path ...string) error {
	filename := ""
	if len(path) > 0 {
		filename = path[0]
	} else {
		filename, _ = ioutil.TempDir(os.TempDir(), "")
	}
	var err error
	db.DB, err = leveldb.OpenFile(filename, nil)
	if err != nil {
		return err
	}
	db.locker = f.NewLocker()
	return nil
}

// Size gets the size of the database in bytes.
func (db *LevelDB) Size() int64 {
	var stats leveldb.DBStats
	if nil != db.DB.Stats(&stats) {
		return -1
	}
	size := int64(0)
	for _, v := range stats.LevelSizes {
		size += v
	}
	return size
}

// Incr increment the key by the specified value.
func (db *LevelDB) Incr(k string, by int64) (int64, error) {
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
func (db *LevelDB) Set(k, v string, ttl int) error {
	return db.set(k, v, ttl)
}

// SetBytes sets a key with the specified value and optional ttl.seconds
func (db *LevelDB) SetBytes(k, v []byte, ttl int) error {
	return db.set(f.String(k), f.String(v), ttl)
}

// MSet sets multiple key-value pairs.
func (db *LevelDB) MSet(data map[string]string) error {
	batch := new(leveldb.Batch)
	for k, v := range data {
		v = "0;" + v
		batch.Put(f.Bytes(k), f.Bytes(v))
	}
	return db.DB.Write(batch, nil)
}

// Get fetches the value of the specified k.
func (db *LevelDB) Get(k string) (string, error) {
	return db.get(k)
}

// MGet fetch multiple values of the specified keys.
func (db *LevelDB) MGet(keys []string) (data []string) {
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
func (db *LevelDB) TTL(key string) int64 {
	item, err := db.DB.Get(f.Bytes(key), nil)
	if err != nil {
		return -2
	}

	parts := strings.SplitN(f.String(item), ";", 2)
	exp, _ := strconv.ParseInt(parts[0], 10, 0)
	if exp == 0 {
		return -1
	}

	now := time.Now().Unix()
	if now >= exp {
		return -2
	}

	return (exp - now) / int64(time.Second)
}

// Del removes key(s) from the store.
func (db *LevelDB) Del(keys []string) error {
	batch := new(leveldb.Batch)
	for _, key := range keys {
		batch.Delete(f.Bytes(key))
	}
	return db.DB.Write(batch, nil)
}

// Close ...
func (db *LevelDB) Close() error {
	return db.DB.Close()
}

// Keys gets matched keys.
func (db *LevelDB) Keys(prefix ...string) []string {
	ro := &opt.ReadOptions{DontFillCache: true}
	var prefixBytes []byte
	var keys []string
	l := len(prefix)
	if l > 0 {
		prefixBytes = f.Bytes(prefix[0])
	}
	var iter iterator.Iterator
	if l == 0 {
		iter = db.DB.NewIterator(nil, ro)
	} else {
		iter = db.DB.NewIterator(util.BytesPrefix(prefixBytes), ro)
	}
	defer iter.Release()
	for iter.Next() {
		if iter.Error() != nil {
			continue
		}
		keys = append(keys, f.String(iter.Key()))
	}
	return keys
}

// GC runs the garbage collector.
func (db *LevelDB) GC() error {
	return db.DB.CompactRange(util.Range{})
}

func (db *LevelDB) get(k string) (string, error) {
	var data string
	var err error
	var del bool

	item, err := db.DB.Get(f.Bytes(k), nil)
	if err != nil {
		return "", err
	}

	parts := strings.SplitN(f.String(item), ";", 2)
	expires, actual := parts[0], parts[1]

	if exp, _ := strconv.ParseInt(expires, 10, 0); exp > 0 && time.Now().Unix() >= exp {
		del = true
		err = errors.New("key not found")
	} else {
		data = actual
	}

	if del {
		_ = db.DB.Delete(f.Bytes(k), nil)
		return data, errors.New("key not found")
	}

	return data, nil
}

func (db *LevelDB) set(k, v string, ttl int) error {
	var expires int64
	if ttl > 0 {
		expires = time.Now().Add(time.Duration(ttl) * time.Second).Unix()
	}
	v = strconv.FormatInt(expires, 10) + ";" + v
	return db.DB.Put(f.Bytes(k), f.Bytes(v), nil)
}
