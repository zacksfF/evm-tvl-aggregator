package api

import (
	"fmt"
	"sync"
	"time"
)

type MemoryCache struct {
	data map[string]*cacheItem
	mu   sync.RWMutex
}

type cacheItem struct {
	value  []byte
	expiry time.Time
}

func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]*cacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

func (c *MemoryCache) Get(key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}

	if time.Now().After(item.expiry) {
		return nil, fmt.Errorf("cache expired")
	}

	return item.value, nil
}

func (c *MemoryCache) Set(key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cacheItem{
		value:  value,
		expiry: time.Now().Add(ttl),
	}

	return nil
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.data {
			if now.After(item.expiry) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}
