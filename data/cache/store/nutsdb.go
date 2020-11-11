package store

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/angenalZZZ/gofunc/f"
	"github.com/xujiajun/nutsdb"
)

// Ndb global nutsdb client
var Ndb *nutsdb.DB

const (
	// NdbType represents the storage type as a string value
	NdbType = "ndb"
	// NdbTagPattern represents the tag pattern to be used as a key in specified storage
	NdbTagPattern = "gocache_tag_%s"
)

// NdbStore is a store for nutsdb
type NdbStore struct {
	// ClientInterface represents github.com/xujiajun/nutsdb client
	client  *nutsdb.DB
	options *Options
	bucket  string
}

// NewNdb creates a new store to nutsdb instance(s)
func NewNdb(option *Options, bucketAndPath ...string) *NdbStore {
	if option == nil {
		option = &Options{}
	}

	var opt, l, bucket = nutsdb.DefaultOptions, len(bucketAndPath), "bucket"
	if l == 1 {
		bucket = bucketAndPath[0]
	}
	if l == 2 {
		opt.Dir = bucketAndPath[1]
	} else {
		opt.Dir = filepath.Join(f.CurrentDir(), ".nutsdb")
	}

	// EntryIdxMode 代表索引entry的模式. 选项: HintKeyValAndRAMIdxMode、HintKeyAndRAMIdxMode和HintBPTSparseIdxMode
	// 其中 HintKeyValAndRAMIdxMode 代表纯内存索引模式（key和value都会被cache）
	// 其中 HintKeyAndRAMIdxMode 代表内存+磁盘的索引模式（只有key被cache）
	// 其中 HintBPTSparseIdxMode 是专门节约内存的设计方案，单机10亿条数据，只要80几M内存。但是读性能不高，需要自己加缓存来加速
	opt.EntryIdxMode = nutsdb.HintKeyValAndRAMIdxMode
	// RWMode 代表读写模式. RWMode 包括两种选项: FileIO and MMap. FileIO 用标准的 I/O读写。 MMap 代表使用mmap进行读写
	opt.RWMode = nutsdb.FileIO
	// SegmentSize 代表数据库的数据单元，每个数据单元（文件）为SegmentSize
	// 现在默认是8 MB，这个可以自己配置。但是一旦被设置，下次启动数据库也要用这个配置，不然会报错
	opt.SegmentSize = 8 * 1024 * 1024
	// NodeNum:1 代表节点的号码，取值范围 [1,1023]
	opt.NodeNum = 1
	// SyncEnable:false 写性能会很高，但是如果遇到断电或者系统奔溃，会有数据丢失的风险
	// SyncEnable:true 写性能会相比false的情况慢很多，但是数据更有保障，每次事务提交成功都会落盘
	opt.SyncEnable = false // ***此选项与DefaultOptions不同***
	// StartFileLoadingMode 代表启动数据库的载入文件的方式。参数选项同RWMode
	opt.StartFileLoadingMode = nutsdb.MMap

	client, err := nutsdb.Open(opt)
	if err != nil {
		return nil
	}

	return &NdbStore{
		client:  client,
		options: option,
		bucket:  bucket,
	}
}

// Get returns data stored from a given key
func (s *NdbStore) Get(key string) (interface{}, error) {
	var data []byte
	err := s.client.View(func(tx *nutsdb.Tx) error {
		item, err := tx.Get(s.bucket, f.Bytes(key))
		if err != nil {
			return err
		}
		data = item.Value
		return nil
	})
	return data, err
}

// TTL returns a expiration time
func (s *NdbStore) TTL(key string) (time.Duration, error) {
	var expires int64

	_ = s.client.View(func(txn *nutsdb.Tx) error {
		item, err := txn.Get(s.bucket, f.Bytes(key))
		if err != nil {
			expires = -2
			return nil
		}

		exp := item.Meta.TTL
		if exp == 0 {
			expires = -1
			return nil
		}

		expires = int64(exp)
		return nil
	})

	if 0 >= expires {
		return -2, errors.New("unable to retrieve data from nutsdb")
	}

	return time.Second * time.Duration(expires), nil
}

// Set defines data in nutsdb for given key identifier
func (s *NdbStore) Set(key string, value interface{}, options *Options) error {
	if options == nil {
		options = s.options
	}

	err := s.client.Update(func(txn *nutsdb.Tx) (err error) {
		if options.Expiration <= 0 {
			err = txn.Put(s.bucket, f.Bytes(key), value.([]byte), 0)
		} else {
			err = txn.Put(s.bucket, f.Bytes(key), value.([]byte), uint32(options.Expiration.Seconds()))
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

func (s *NdbStore) setTags(key string, tags []string) {
	for _, tag := range tags {
		var tagKey = fmt.Sprintf(NdbTagPattern, tag)
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

// Delete removes data from nutsdb for given key identifier
func (s *NdbStore) Delete(key string) error {
	return s.client.Update(func(txn *nutsdb.Tx) error {
		return txn.Delete(s.bucket, f.Bytes(key))
	})
}

// Invalidate invalidates some cache data in nutsdb for given options
func (s *NdbStore) Invalidate(options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			var tagKey = fmt.Sprintf(NdbTagPattern, tag)
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

// Clear resets all data in the store.
// 随着数据越来越多，特别是一些删除或者过期的数据占据着磁盘，
// 清理这些NutsDB提供了db.Merge()方法，这个方法需要自己根据实际情况编写合并策略。
// 一旦执行会影响到正常的写请求，所以最好避开高峰期，比如半夜定时执行等。
func (s *NdbStore) Clear() error {
	return s.client.Merge()
}

// Backup backup all data in the dir.
// 数据库的备份。这个方法执行的是一个热备份，不会阻塞到数据库其他的只读事务操作，对写事务会有影响。
func (s *NdbStore) Backup(dir string) error {
	return s.client.Backup(dir)
}

// GetType returns the store type
func (s *NdbStore) GetType() string {
	return NdbType
}
