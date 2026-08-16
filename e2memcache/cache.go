package e2memcache

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type Cache[T any] struct {
	mu          sync.RWMutex
	m           map[string]T
	lastUpdated time.Time
}

func New[T any]() *Cache[T] {
	return &Cache[T]{
		m:  make(map[string]T),
		mu: sync.RWMutex{},
	}
}

func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
	c.lastUpdated = time.Now()
}

func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[key]
	return v, ok
}

func (c *Cache[T]) Del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, key)
	c.lastUpdated = time.Now()
}

func (c *Cache[T]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.m)
}

func (c *Cache[T]) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]string, 0, len(c.m))
	for k := range c.m {
		keys = append(keys, k)
	}
	return keys
}

func (c *Cache[T]) Values() []T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	values := make([]T, 0, len(c.m))
	for _, v := range c.m {
		values = append(values, v)
	}
	return values
}

func (c *Cache[T]) LastUpdated() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastUpdated
}

func (c *Cache[T]) LastUpdatedString() string {
	return c.lastUpdated.String()
}

func (c *Cache[T]) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	clear(c.m)
	c.lastUpdated = time.Now()
}

func (c *Cache[T]) Filter(f func(T) bool) *Cache[T] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := New[T]()
	for k, v := range c.m {
		if f(v) {
			result.m[k] = v
		}
	}
	result.lastUpdated = time.Now()
	return result
}

func (c *Cache[T]) FilterValues(f func(T) bool) []T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	values := make([]T, 0, len(c.m)/4)
	for _, v := range c.m {
		if f(v) {
			values = append(values, v)
		}
	}
	return values
}

func (c *Cache[T]) Dump(out io.Writer) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for k, v := range c.m {
		_, _ = fmt.Fprintf(out, "%v: %v\n", k, v)
	}
}
