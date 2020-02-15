package memcache

import "time"

// Get - fetch item from cache with ok indicating status of the operation
func (mc *CacheType) Get(key string) (value interface{}, ok bool) {
	mc.m.RLock()
	defer mc.m.RUnlock()
	valueType, ok := mc.cache[key]

	if ok {
		value = valueType.Value
	}

	return
}

// GetByID - fetch item from cache based on item ID. Return nil otherwise
func (mc *CacheType) GetByID(itemID int64) (item interface{}) {
	mc.m.RLock()
	defer mc.m.RUnlock()

	if itemID < 0 {
		return nil
	}

	if itemID <= mc.Len() {
		item = mc.items[itemID-1]
	}

	return
}

// Add - add item to cache and return resulting item ID
func (mc *CacheType) Add(key string, value interface{}) (itemID int64) {
	if _, ok := mc.Get(key); ok {
		return
	}

	mc.m.Lock()
	defer mc.m.Unlock()

	valueStruct := &ValueType{
		Value:    value,
		Expires:  0,
		MetaData: key,
	}

	mc.items = append(mc.items, valueStruct)
	mc.cache[key] = valueStruct

	return mc.Len()
}

// Set - set key/value pair without expiration
func (mc *CacheType) Set(key string, value interface{}) {
	mc.SetEx(key, value, 0)
}

// SetEx - set key/value pair with expiration
func (mc *CacheType) SetEx(key string, value interface{}, expires int64) {
	mc.m.Lock()
	defer mc.m.Unlock()

	if expires > 0 {
		expires += time.Now().Unix()
	}

	mc.cache[key] = &ValueType{
		Value:   value,
		Expires: expires,
	}
}

// Len - returns cache length
func (mc *CacheType) Len() (cacheSize int64) {
	cacheSize = int64(len(mc.cache))
	return
}

// LenSafe - same as Len() but with read lock
func (mc *CacheType) LenSafe() (cacheSize int64) {
	mc.m.RLock()
	defer mc.m.RUnlock()

	return mc.Len()
}

// Cache - return whole cache contents
func (mc *CacheType) Cache() (cache map[string]*ValueType) {
	cache = mc.cache
	return
}

// UnsafeDelete - removes item from cache
func (mc *CacheType) UnsafeDelete(key string) {
	delete(mc.cache, key)

	idx := 0

	for valuePos, value := range mc.items {
		if value.MetaData == key {
			idx = valuePos
			break
		}
	}
	// I should really really test this :)  - SliceTricks
	if idx < len(mc.items)-1 {
		copy(mc.items[idx:], mc.items[idx+1:])
	}
	if len(mc.items) >= 1 {
		mc.items[len(mc.items)-1] = nil
		mc.items = mc.items[:len(mc.items)-1]
	}
}

// Delete - removes item from cache
func (mc *CacheType) Delete(key string) {
	mc.m.Lock()
	defer mc.m.Unlock()

	mc.UnsafeDelete(key)
}

// Evictor - background goroutine that periodically purges expired keys
func (mc *CacheType) Evictor() {
	for {
		select {
		case <-mc.done:
			return
		case <-mc.ticker.C:
			mc.m.Lock()
			for key, value := range mc.cache {
				if value.Expires == 0 {
					continue
				}

				if value.Expires-time.Now().Unix() <= 0 {
					mc.logger.Printf("Evicting %s\n", key)
					mc.UnsafeDelete(key)
				}
			}
			mc.m.Unlock()
		}
	}
}

// Stop - run some cleaning chores during shutdown
func (mc *CacheType) Stop() {
	mc.ticker.Stop()
	mc.done <- struct{}{}

	mc.logger.Debug("Memcache is saying goodbye!")
}

// New - prepare and populate memcache instance
func New(logger Logger) (memCache *CacheType) {
	memCache = &CacheType{cache: make(map[string]*ValueType),
		done:   make(chan struct{}),
		ticker: time.NewTicker(1 * time.Second),
		logger: logger,
	}
	go memCache.Evictor()

	return
}
