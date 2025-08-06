package pokecache

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	mu  sync.RWMutex
	cache map[string]cacheEntry
	ttl   time.Duration
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(ttl time.Duration) *Cache {
	newCache := &Cache{
		cache: make(map[string]cacheEntry),
		ttl:   ttl,
	}

	// Run cleanup every 1/4 of TTL, minimum 1 minute
	cleanupInterval := ttl / 4
	if cleanupInterval < time.Minute {
		cleanupInterval = time.Minute
	}
	
	go newCache.reapLoop(cleanupInterval)

	return newCache
}

func (cacheData *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		cacheData.mu.Lock()
		for key, entry := range cacheData.cache {
			if time.Since(entry.createdAt) > cacheData.ttl {
				delete(cacheData.cache, key)
			}
		}
		cacheData.mu.Unlock()
	}
}

func (cacheData *Cache) Get(key string) ([]byte, bool) {
	if key == "" {
		return nil, false
	}

	cacheData.mu.RLock()
	entry, exists := cacheData.cache[key]
	cacheData.mu.RUnlock()

	if !exists {
		return nil, false
	}

	return entry.val, true
}

func (cacheData *Cache) Add(key string, val []byte) error {
	if key == "" || val == nil {
		return fmt.Errorf("key and value must not be empty")
	}

	cacheData.mu.Lock()
	cacheData.cache[key] = cacheEntry{createdAt: time.Now(), val: val}
	cacheData.mu.Unlock()

	return nil
}
