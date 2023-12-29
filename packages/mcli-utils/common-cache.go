package mcliutils

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

var testingkey = "testing4893472"

type wrapCacheEntry struct {
	key string
	e   *CacheEntry
}
type CacheEntry struct {
	mu         sync.Mutex // guards
	Value      interface{}
	UpdateFunc CacheFunc
	Ttl        time.Duration
	TimeStamp  time.Time
	Count      uint
}

type CCache struct {
	Func       CacheFunc
	mu         sync.RWMutex
	Cache      map[string]*CacheEntry
	Ttl        time.Duration // time in milliseconds to store entry in cache
	MaxEntries int           //store max number of entries in Optimized cache if <=0 --> no limits
}

type CacheFunc func(params ...interface{}) (interface{}, error)

func NewCCache(ttl time.Duration, maxEntries int, f CacheFunc, ctx context.Context, notify chan interface{}) *CCache {
	return &CCache{Ttl: ttl, MaxEntries: maxEntries, Cache: make(map[string]*CacheEntry), Func: f}
}

type CCacheEntries []wrapCacheEntry

func (c CCacheEntries) Len() int      { return len(c) }
func (c CCacheEntries) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c CCacheEntries) Less(i, j int) bool {
	return c[i].e.Count*uint(c[i].e.TimeStamp.Unix()) < c[j].e.Count*uint(c[j].e.TimeStamp.Unix())
}

func (cc *CCache) Get(key string) (interface{}, error) {
	var entry *CacheEntry

	cc.mu.RLock()
	entry, ok := cc.Cache[key]
	cc.mu.RUnlock()
	if ok {
		entry.mu.Lock()
		entry.Count++
		entry.mu.Unlock()
		return entry.Value, nil
	}
	return nil, fmt.Errorf("key %s doesn't exists", key)
}

func (cc *CCache) GetAndSetIfNotExists(key string, value ...interface{}) (interface{}, error) {
	var entry *CacheEntry
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
		entry = &CacheEntry{Count: 1}
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
	if len(cc.Cache) > int(float64(cc.MaxEntries)*1.2) {
		cc.Optimize(true)
	}
	cc.mu.Unlock()

	return entry.Value, nil
}

func (cc *CCache) Set(key string, updateFunc func(params ...interface{}) (interface{}, error), ttl time.Duration, value ...interface{}) (interface{}, error) {
	var entry *CacheEntry = &CacheEntry{Count: 1, UpdateFunc: updateFunc, Ttl: ttl, TimeStamp: time.Now()}
	var err error

	if len(value) == 0 {
		return nil, fmt.Errorf("no value provided to store in cache")
	}
	if updateFunc != nil {
		entry.Value, err = updateFunc(value...)
		if err != nil {
			return nil, err
		}
	} else if cc.Func != nil {
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
	if len(cc.Cache) > int(float64(cc.MaxEntries)*1.2) {
		cc.Optimize(true)
	}
	// fmt.Println(cc.Cache)
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

func (cc *CCache) Optimize(inLockMode bool) bool {

	ok := false
	currentTime := time.Now()
	limit := int(float64(cc.MaxEntries) * 1.2)
	wrapCacheEntries := make(CCacheEntries, 0, len(cc.Cache))
	if cc.Ttl > 0 || cc.MaxEntries > 0 {
		for key, entry := range cc.Cache {
			expirationTime := entry.TimeStamp.Add(time.Millisecond * time.Duration(cc.Ttl))
			if currentTime.After(expirationTime) {
				if inLockMode {
					delete(cc.Cache, key)
				} else {
					cc.Remove(key)
				}
				ok = true
				continue
			}
			wrapCacheEntries = append(wrapCacheEntries, wrapCacheEntry{key: key, e: entry})
		}
	}
	// if we need save only max usable entries
	if cc.MaxEntries > 0 && len(cc.Cache) >= limit {
		sort.Sort(sort.Reverse(wrapCacheEntries))
		_ = limit
		for idx, e := range wrapCacheEntries {
			if (idx + 1) > cc.MaxEntries {
				if inLockMode {
					delete(cc.Cache, e.key)
				} else {
					cc.Remove(e.key)
				}
				ok = true
			}
		}
		// for key, entry := range cc.Cache {
		// 	fmt.Println(key, entry.Value, entry.TimeStamp, entry.Count)
		// }
	}
	return ok
}
