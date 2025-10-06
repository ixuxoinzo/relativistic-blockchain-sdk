package core

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Cache struct {
	items    map[string]*CacheItem
	mu       sync.RWMutex
	logger   *zap.Logger
	ttl      time.Duration
	maxSize  int
}

type CacheItem struct {
	Value      interface{}
	ExpiresAt  time.Time
	CreatedAt  time.Time
	AccessCount int
}

func NewCache(ttl time.Duration, maxSize int, logger *zap.Logger) *Cache {
	return &Cache{
		items:   make(map[string]*CacheItem),
		logger:  logger,
		ttl:     ttl,
		maxSize: maxSize,
	}
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	c.items[key] = &CacheItem{
		Value:      value,
		ExpiresAt:  time.Now().Add(c.ttl),
		CreatedAt:  time.Now(),
		AccessCount: 1,
	}

	c.logger.Debug("Cache item set", zap.String("key", key))
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}

	c.mu.Lock()
	item.AccessCount++
	c.mu.Unlock()

	return item.Value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	c.logger.Debug("Cache item deleted", zap.String("key", key))
}

func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
		c.logger.Debug("Evicted oldest cache item", zap.String("key", oldestKey))
	}
}

func (c *Cache) CleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	count := 0

	for key, item := range c.items {
		if now.After(item.ExpiresAt) {
			delete(c.items, key)
			count++
		}
	}

	if count > 0 {
		c.logger.Debug("Cleaned up expired cache items", zap.Int("count", count))
	}
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
	c.logger.Info("Cache cleared")
}

func (c *Cache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{
		TotalItems: len(c.items),
	}

	now := time.Now()
	for _, item := range c.items {
		stats.TotalAccessCount += item.AccessCount
		if now.After(item.ExpiresAt) {
			stats.ExpiredItems++
		}
	}

	return stats
}

type CacheStats struct {
	TotalItems       int
	ExpiredItems     int
	TotalAccessCount int
}