package f

import "sync"

// CiMapShardCount Map Shard Count
var CiMapShardCount = 32

// CiMap thread safe map of type uint64:Anything.
// To avoid lock bottlenecks this map is dived to several (SHARD_COUNT) map shards.
type CiMap []*CiMapShared

// CiMapShared thread safe uint64 to anything map.
type CiMapShared struct {
	items        map[uint64]interface{}
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

// NewCiMap Creates a new concurrent map.
func NewCiMap() CiMap {
	m := make(CiMap, CiMapShardCount)
	for i := 0; i < CiMapShardCount; i++ {
		m[i] = &CiMapShared{items: make(map[uint64]interface{})}
	}
	return m
}

// NewCiMapFromJSON Creates a new concurrent map.
func NewCiMapFromJSON(json []byte) (m CiMap, err error) {
	tmp := make(map[uint64]interface{})
	err = DecodeJson(json, &tmp)
	if err == nil {
		m = NewCiMap()
		m.MSet(tmp)
	}
	return
}

// GetShard returns shard under given key
func (m CiMap) GetShard(key uint64) *CiMapShared {
	return m[key%uint64(CiMapShardCount)]
}

// MSet sets values.
func (m CiMap) MSet(data map[uint64]interface{}) {
	for key, value := range data {
		shard := m.GetShard(key)
		shard.Lock()
		shard.items[key] = value
		shard.Unlock()
	}
}

// Set the given value under the specified key.
func (m CiMap) Set(key uint64, value interface{}) {
	// GetHeader map shard.
	shard := m.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

// Up Insert or Update - updates existing element or inserts a new one using CMapUpCb
func (m CiMap) Up(key uint64, value interface{}, cb CMapUpCb) (res interface{}) {
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	res = cb(ok, v, value)
	shard.items[key] = res
	shard.Unlock()
	return res
}

// SetIfAbsent Sets the given value under the specified key if no value was associated with it.
func (m CiMap) SetIfAbsent(key uint64, value interface{}) bool {
	// GetHeader map shard.
	shard := m.GetShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return !ok
}

// Get retrieves an element from map under given key.
func (m CiMap) Get(key uint64) (interface{}, bool) {
	// GetHeader shard
	shard := m.GetShard(key)
	shard.RLock()
	// GetHeader item from shard.
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

// Count returns the number of elements within the map.
func (m CiMap) Count() int {
	count := 0
	for i := 0; i < CiMapShardCount; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Has Looks up an item under specified key
func (m CiMap) Has(key uint64) bool {
	// GetHeader shard
	shard := m.GetShard(key)
	shard.RLock()
	// See if element is within shard.
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

// Remove removes an element from the map.
func (m CiMap) Remove(key uint64) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

// RemoveCbi is a callback executed in a map.RemoveCb() call, while Lock is held
// If returns true, the element will be removed from the map
type RemoveCbi func(key uint64, v interface{}, exists bool) bool

// RemoveCb locks the shard containing the key, retrieves its current value and calls the callback with those params
// If callback returns true and element exists, it will remove it from the map
// Returns the value returned by the callback (even if element was not present in the map)
func (m CiMap) RemoveCb(key uint64, cb RemoveCbi) bool {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	remove := cb(key, v, ok)
	if remove && ok {
		delete(shard.items, key)
	}
	shard.Unlock()
	return remove
}

// Pop removes an element from the map and returns it
func (m CiMap) Pop(key uint64) (v interface{}, exists bool) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	v, exists = shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return v, exists
}

// IsEmpty checks if map is empty.
func (m CiMap) IsEmpty() bool {
	return m.Count() == 0
}

// TupleI Used by the Iter & IterBuffered functions to wrap two variables together over a channel
type TupleI struct {
	Key uint64
	Val interface{}
}

// Iter returns an iterator which could be used in a for range loop.
//
// Deprecated: using IterBuffered() will get a better performence
func (m CiMap) Iter() <-chan TupleI {
	chans := snapshotI(m)
	ch := make(chan TupleI)
	go fanInI(chans, ch)
	return ch
}

// IterBuffered returns a buffered iterator which could be used in a for range loop.
func (m CiMap) IterBuffered() <-chan TupleI {
	chans := snapshotI(m)
	total := 0
	for _, c := range chans {
		total += cap(c)
	}
	ch := make(chan TupleI, total)
	go fanInI(chans, ch)
	return ch
}

// Returns a array of channels that contains elements in each shard,
// which likely takes a snapshot of `m`.
// It returns once the size of each buffered channel is determined,
// before all the channels are populated using goroutines.
func snapshotI(m CiMap) (chans []chan TupleI) {
	chans = make([]chan TupleI, CiMapShardCount)
	wg := sync.WaitGroup{}
	wg.Add(CiMapShardCount)
	// Foreach shard.
	for index, shard := range m {
		go func(index int, shard *CiMapShared) {
			// Foreach key, value pair.
			shard.RLock()
			chans[index] = make(chan TupleI, len(shard.items))
			wg.Done()
			for key, val := range shard.items {
				chans[index] <- TupleI{key, val}
			}
			shard.RUnlock()
			close(chans[index])
		}(index, shard)
	}
	wg.Wait()
	return chans
}

// fanInI reads elements from channels `chans` into channel `out`
func fanInI(chans []chan TupleI, out chan TupleI) {
	wg := sync.WaitGroup{}
	wg.Add(len(chans))
	for _, ch := range chans {
		go func(ch chan TupleI) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(out)
}

// Items returns all items as map[uint64]interface{}
func (m CiMap) Items() map[uint64]interface{} {
	tmp := make(map[uint64]interface{})

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}

	return tmp
}

// IterCbi Iterator callback,called for every key,value found in
// maps. RLock is held for all calls for a given shard
// therefore callback sess consistent view of a shard,
// but not across the shards
type IterCbi func(key uint64, v interface{})

// IterCb Callback based iterator, cheapest way to read
// all elements in a map.
func (m CiMap) IterCb(fn IterCbi) {
	for idx := range m {
		shard := (m)[idx]
		shard.RLock()
		for key, value := range shard.items {
			fn(key, value)
		}
		shard.RUnlock()
	}
}

// Keys returns all keys as []uint64
func (m CiMap) Keys() []uint64 {
	count := m.Count()
	ch := make(chan uint64, count)
	go func() {
		// Foreach shard.
		wg := sync.WaitGroup{}
		wg.Add(CiMapShardCount)
		for _, shard := range m {
			go func(shard *CiMapShared) {
				// Foreach key, value pair.
				shard.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.RUnlock()
				wg.Done()
			}(shard)
		}
		wg.Wait()
		close(ch)
	}()

	// Generate keys
	keys := make([]uint64, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}

// JSON Reviles CiMap "private" variables to json marshal.
func (m CiMap) JSON() ([]byte, error) {
	tmp := make(map[uint64]interface{})

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}
	return EncodeJson(tmp)
}
