package cache

import (
	"sync"
	"time"
)

type Entry[T any] struct {
	Value     T
	ExpiresAt time.Time
}
type Cache[T any] struct {
	mu    sync.RWMutex
	items map[string]Entry[T]
	ttl   time.Duration
}

func New[T any](ttl time.Duration) *Cache[T] {
	return &Cache[T]{items: map[string]Entry[T]{}, ttl: ttl}
}
func (c *Cache[T]) Set(key string, value T, now time.Time) {
	c.mu.Lock()
	c.items[key] = Entry[T]{Value: value, ExpiresAt: now.Add(c.ttl)}
	c.mu.Unlock()
}
func (c *Cache[T]) Get(key string, now time.Time) (T, bool) {
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || !now.Before(entry.ExpiresAt) {
		var zero T
		return zero, false
	}
	return entry.Value, true
}
func (c *Cache[T]) Delete(key string) { c.mu.Lock(); delete(c.items, key); c.mu.Unlock() }
func (c *Cache[T]) Purge(now time.Time) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for key, entry := range c.items {
		if !now.Before(entry.ExpiresAt) {
			delete(c.items, key)
			n++
		}
	}
	return n
}
func (c *Cache[T]) Len() int { c.mu.RLock(); defer c.mu.RUnlock(); return len(c.items) }
