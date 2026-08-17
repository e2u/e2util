package e2gocache

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultExpiration time.Duration = -1
)

type Cache[T any] struct {
	mu          sync.RWMutex
	m           map[string]*storage[T]
	ttl         time.Duration
	lastUpdated atomic.Int64

	stopCleaner chan struct{}
	stopOnce    sync.Once
}

type storage[T any] struct {
	s       T
	ttl     time.Duration // <=0 表示永不過期
	expires atomic.Int64  // UnixNano；0=永不過期
}

func New[T any](expiration time.Duration) *Cache[T] {
	c := &Cache[T]{
		m:           make(map[string]*storage[T]),
		mu:          sync.RWMutex{},
		ttl:         DefaultExpiration,
		stopCleaner: make(chan struct{}),
	}
	if expiration > 0 {
		c.ttl = expiration
	}
	c.startCleaner(time.Second)
	return c
}

func (c *Cache[T]) Set(key string, value T, expiration ...time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ttl := c.ttl
	if len(expiration) > 0 {
		ttl = expiration[0]
	}

	now := time.Now()
	item := &storage[T]{s: value, ttl: ttl}
	if ttl > 0 {
		item.expires.Store(now.Add(ttl).UnixNano())
	}
	c.m[key] = item
	c.lastUpdated.Store(now.UnixNano())
}

func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var zero T
	v, ok := c.m[key]
	if !ok {
		return zero, false
	}
	if exp := v.expires.Load(); exp > 0 && time.Now().UnixNano() >= exp {
		return zero, false // 或升級為刪除
	}
	return v.s, true
}

func (c *Cache[T]) Del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, key)
	c.lastUpdated.Store(time.Now().UnixNano())
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
		values = append(values, v.s)
	}
	return values
}

func (c *Cache[T]) LastUpdated() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Unix(0, c.lastUpdated.Load())
}

func (c *Cache[T]) LastUpdatedString() string {
	return c.LastUpdated().UTC().Format(time.RFC3339Nano)
}

func (c *Cache[T]) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	clear(c.m)
	c.lastUpdated.Store(time.Now().UnixNano())
}

func (c *Cache[T]) Filter(f func(T) bool, expiration ...time.Duration) *Cache[T] {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ttl := c.ttl
	if len(expiration) > 0 {
		ttl = expiration[0]
	}
	result := &Cache[T]{
		m:   make(map[string]*storage[T]),
		ttl: ttl,
		mu:  sync.RWMutex{},
	}

	now := time.Now()
	for k, v := range c.m {
		if exp := v.expires.Load(); exp > 0 && now.UnixNano() >= exp {
			continue
		}
		if !f(v.s) {
			continue
		}
		cp := &storage[T]{s: v.s, ttl: ttl}
		if ttl > 0 {
			cp.expires.Store(now.Add(ttl).UnixNano())
		}
		result.m[k] = cp
	}
	result.lastUpdated.Store(time.Now().UnixNano())
	return result
}

func (c *Cache[T]) FilterValues(f func(T) bool) []T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	values := make([]T, 0, len(c.m)/4)
	for _, v := range c.m {
		if f(v.s) {
			values = append(values, v.s)
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

func (c *Cache[T]) cleanupExpired() {
	c.mu.RLock()
	now := time.Now().UnixNano()
	expiredKeys := make([]string, 0, len(c.m))
	for k := range c.m {
		if v, ok := c.m[k]; ok && v.ttl > 0 && v.expires.Load() > 0 && now >= v.expires.Load() {
			expiredKeys = append(expiredKeys, k)
		}
	}
	c.mu.RUnlock()

	if len(expiredKeys) == 0 {
		return
	}

	c.mu.Lock()
	for _, k := range expiredKeys {
		if v, ok := c.m[k]; ok && v.ttl > 0 && v.expires.Load() > 0 && now >= v.expires.Load() {
			delete(c.m, k)
		}
	}
	c.mu.Unlock()
}

func (c *Cache[T]) startCleaner(every time.Duration) {
	if every <= 0 {
		every = time.Minute
	}
	go func() {
		t := time.NewTicker(every)
		defer t.Stop()
		for {
			select {
			case <-c.stopCleaner:
				return
			case <-t.C:
				c.cleanupExpired()
			}
		}
	}()
}

func (c *Cache[T]) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopCleaner)
	})
}
