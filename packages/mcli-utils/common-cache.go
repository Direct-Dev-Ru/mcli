package mcliutils

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

var testingkey = "testing4893472"

type wrapCacheEntry struct {
	key string
	e   *cacheEntry
}
type cacheEntry struct {
	Value     interface{}
	TimeStamp time.Time
	mu        sync.Mutex // guards
	Count     uint
}

type CCache struct {
	Func       CacheFunc
	mu         sync.RWMutex
	Cache      map[string]*cacheEntry
	Ttl        int // time in milliseconds to store entry in cache
	MaxEntries int //store max number of entries in Optimized cache if <=0 --> no limits
}

type CacheFunc func(params ...interface{}) (interface{}, error)

func NewCCache(ttl int, f CacheFunc) *CCache {
	return &CCache{Ttl: ttl, Cache: make(map[string]*cacheEntry), Func: f}
}

type CCacheEntries []wrapCacheEntry

func (c CCacheEntries) Len() int      { return len(c) }
func (c CCacheEntries) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c CCacheEntries) Less(i, j int) bool {
	return c[i].e.Count*uint(c[i].e.TimeStamp.Unix()) < c[j].e.Count*uint(c[j].e.TimeStamp.Unix())
}

func (cc *CCache) GetAndSetIfNotExists(key string, value ...interface{}) (interface{}, error) {
	var entry *cacheEntry
	var err error
	cc.mu.RLock()
	entry, ok := cc.Cache[key]
	cc.mu.RUnlock()
	if ok {
		entry.mu.Lock()
		entry.Count++
		if key == testingkey { //for testing parallelizm
			fmt.Println("i'm sleeping for 2 seconds")
			time.Sleep(2 * time.Second)
		}
		entry.mu.Unlock()
		return entry.Value, nil
	}

	// if key does not exist - process new value
	if !ok {
		if len(value) == 0 {
			return nil, fmt.Errorf("no value provided to store in cache")
		}
		entry = &cacheEntry{Count: 1}
		if cc.Func != nil {
			entry.Value, err = cc.Func(value...)
			if err != nil {
				return nil, err
			}
		} else {
			if len(value) == 1 {
				entry.Value = value[0]
			} else {
				entry.Value = value
			}
		}
	}

	cc.mu.Lock()
	entry.TimeStamp = time.Now()
	cc.Cache[key] = entry
	cc.mu.Unlock()

	return entry.Value, nil
}

func (cc *CCache) Set(key string, value ...interface{}) (interface{}, error) {
	var entry *cacheEntry = &cacheEntry{Count: 1, TimeStamp: time.Now()}
	var err error

	if len(value) == 0 {
		return nil, fmt.Errorf("no value provided to store in cache")
	}

	if cc.Func != nil {
		entry.Value, err = cc.Func(value...)
		if err != nil {
			return nil, err
		}
	} else {
		if len(value) == 1 {
			entry.Value = value[0]
		} else {
			entry.Value = value
		}
	}

	cc.mu.Lock()
	cc.Cache[key] = entry
	cc.mu.Unlock()

	return entry.Value, nil
}

func (cc *CCache) Remove(key string) error {
	if _, exists := cc.Cache[key]; !exists {
		return fmt.Errorf("key '%s' does not exist in cache", key)
	}
	cc.mu.Lock()
	defer cc.mu.Unlock()

	delete(cc.Cache, key)
	return nil
}

func (cc *CCache) Optimize() bool {

	ok := false
	currentTime := time.Now()
	limit := int(float64(cc.MaxEntries) * 1.2)
	wrapCacheEntries := make(CCacheEntries, 0, len(cc.Cache))
	if cc.Ttl > 0 || cc.MaxEntries > 0 {
		for key, entry := range cc.Cache {
			expirationTime := entry.TimeStamp.Add(time.Millisecond * time.Duration(cc.Ttl))
			if currentTime.After(expirationTime) {
				cc.Remove(key)
				ok = true
				continue
			}
			wrapCacheEntries = append(wrapCacheEntries, wrapCacheEntry{key: key, e: entry})
		}
	}
	// if we need save only max usable enries
	if cc.MaxEntries > 0 {
		sort.Sort(sort.Reverse(wrapCacheEntries))
		_ = limit
		for idx, e := range wrapCacheEntries {
			if limit < (idx + 1) {
				cc.Remove(e.key)
				ok = true
			}
		}
		// for key, entry := range cc.Cache {
		// 	fmt.Println(key, entry.Value, entry.TimeStamp, entry.Count)
		// }
	}
	return ok
}
