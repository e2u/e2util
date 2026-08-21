package e2gocache

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultExpiration      time.Duration = -1
	DefaultCleanupInterval               = time.Second
	minCleanupInterval                   = 10 * time.Millisecond
)

type Cache[T any] struct {
	mu           sync.RWMutex
	m            map[string]storage[T]
	ttl          time.Duration
	cleanupEvery time.Duration
	maxKeys      int // 0 = unlimited
	lastUpdated  atomic.Int64
	janitorOn    atomic.Bool
	stopCleaner  chan struct{}
	cleanerOnce  sync.Once
	stopOnce     sync.Once
}

type storage[T any] struct {
	s       T
	ttl     time.Duration // <=0 表示永不過期
	expires int64         // UnixNano；0=永不過期
}

func isExpired[T any](v storage[T], now int64) bool {
	return v.expires > 0 && now >= v.expires
}

func cleanupInterval(expiration time.Duration, override ...time.Duration) time.Duration {
	if len(override) > 0 && override[0] > 0 {
		return override[0]
	}
	if expiration > 0 && expiration < DefaultCleanupInterval {
		if expiration < minCleanupInterval {
			return minCleanupInterval
		}
		return expiration
	}
	return DefaultCleanupInterval
}

// New creates a cache. expiration is the default per-key TTL (<=0 means never).
// An optional second duration is the janitor period. If omitted, it is 1s, or
// the default TTL when that TTL is shorter than 1s (floored at 10ms).
// The janitor goroutine starts on first expiring entry (or immediately if
// expiration > 0).
func New[T any](expiration time.Duration, cleanup ...time.Duration) *Cache[T] {
	c := &Cache[T]{
		m:            make(map[string]storage[T]),
		ttl:          DefaultExpiration,
		cleanupEvery: cleanupInterval(expiration, cleanup...),
		stopCleaner:  make(chan struct{}),
	}
	if expiration > 0 {
		c.ttl = expiration
		c.ensureCleaner()
	}
	return c
}

func (c *Cache[T]) ensureCleaner() {
	if c == nil || c.stopCleaner == nil {
		return
	}
	c.cleanerOnce.Do(func() {
		c.janitorOn.Store(true)
		c.startCleaner(c.cleanupEvery)
	})
}

// WithMaxKeys limits live+stored entries. 0 means unlimited. When inserting a
// new key over the limit, expired keys are dropped first, then one arbitrary key.
func (c *Cache[T]) WithMaxKeys(n int) *Cache[T] {
	if c == nil {
		return c
	}
	c.mu.Lock()
	c.maxKeys = n
	c.dropExpiredLocked(time.Now().UnixNano())
	if n > 0 {
		c.trimToLocked(n)
	}
	c.mu.Unlock()
	return c
}

func (c *Cache[T]) dropExpiredLocked(now int64) {
	for k, v := range c.m {
		if isExpired(v, now) {
			delete(c.m, k)
		}
	}
}

func (c *Cache[T]) trimToLocked(limit int) {
	if limit < 0 {
		limit = 0
	}
	for len(c.m) > limit {
		for k := range c.m {
			delete(c.m, k)
			break
		}
	}
}

func (c *Cache[T]) setLocked(key string, value T, ttl time.Duration, now time.Time) {
	if ttl > 0 {
		c.ensureCleaner()
	}
	_, exists := c.m[key]
	if !exists {
		c.dropExpiredLocked(now.UnixNano())
		if c.maxKeys > 0 {
			c.trimToLocked(c.maxKeys - 1)
		}
	}
	item := storage[T]{s: value, ttl: ttl}
	if ttl > 0 {
		item.expires = now.Add(ttl).UnixNano()
	}
	c.m[key] = item
	c.lastUpdated.Store(now.UnixNano())
}

func (c *Cache[T]) Set(key string, value T, expiration ...time.Duration) {
	ttl := c.ttl
	if len(expiration) > 0 {
		ttl = expiration[0]
	}
	if ttl > 0 {
		c.ensureCleaner()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setLocked(key, value, ttl, time.Now())
}

func (c *Cache[T]) Get(key string) (T, bool) {
	var zero T
	now := time.Now().UnixNano()

	c.mu.RLock()
	v, ok := c.m[key]
	if !ok {
		c.mu.RUnlock()
		return zero, false
	}
	if !isExpired(v, now) {
		val := v.s
		c.mu.RUnlock()
		return val, true
	}
	c.mu.RUnlock()

	c.mu.Lock()
	if v, ok := c.m[key]; ok && isExpired(v, time.Now().UnixNano()) {
		delete(c.m, key)
		c.lastUpdated.Store(time.Now().UnixNano())
	}
	c.mu.Unlock()
	return zero, false
}

// GetOrSet returns the existing value, or stores and returns fn() on miss.
// fn runs under the write lock so only one caller computes the value.
func (c *Cache[T]) GetOrSet(key string, fn func() T, expiration ...time.Duration) T {
	if v, ok := c.Get(key); ok {
		return v
	}
	ttl := c.ttl
	if len(expiration) > 0 {
		ttl = expiration[0]
	}
	if ttl > 0 {
		c.ensureCleaner()
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	if v, ok := c.m[key]; ok && !isExpired(v, now.UnixNano()) {
		return v.s
	}
	val := fn()
	c.setLocked(key, val, ttl, now)
	return val
}

// GetAS fetches key and type-asserts the value to Out (Go 1.27 generic method).
func (c *Cache[S]) GetAS[Out any](key string) (Out, bool) {
	var zero Out
	if c == nil {
		return zero, false
	}
	v, ok := c.Get(key)
	if !ok {
		return zero, false
	}
	out, ok := any(v).(Out)
	return out, ok
}

// GetAS is the package-level form of Cache.GetAS.
func GetAS[Out any, S any](c *Cache[S], key string) (Out, bool) {
	return c.GetAS[Out](key)
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
	now := time.Now().UnixNano()
	n := 0
	for _, v := range c.m {
		if !isExpired(v, now) {
			n++
		}
	}
	return n
}

func (c *Cache[T]) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now().UnixNano()
	keys := make([]string, 0, len(c.m))
	for k, v := range c.m {
		if !isExpired(v, now) {
			keys = append(keys, k)
		}
	}
	return keys
}

func (c *Cache[T]) Values() []T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now().UnixNano()
	values := make([]T, 0, len(c.m))
	for _, v := range c.m {
		if !isExpired(v, now) {
			values = append(values, v.s)
		}
	}
	return values
}

func (c *Cache[T]) LastUpdated() time.Time {
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

type liveKV[T any] struct {
	k string
	v storage[T]
}

// snapshotLive copies unexpired entries under RLock, then returns so callers
// (Filter / FilterValues) can run predicates without holding the cache lock.
func (c *Cache[T]) snapshotLive() ([]liveKV[T], int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now().UnixNano()
	out := make([]liveKV[T], 0, len(c.m))
	for k, v := range c.m {
		if isExpired(v, now) {
			continue
		}
		out = append(out, liveKV[T]{k: k, v: v})
	}
	return out, c.maxKeys
}

func (c *Cache[T]) Filter(f func(T) bool, expiration ...time.Duration) *Cache[T] {
	items, maxKeys := c.snapshotLive()

	override := len(expiration) > 0
	ttl := c.ttl
	if override {
		ttl = expiration[0]
	}

	every := c.cleanupEvery
	if override {
		every = cleanupInterval(ttl)
	}
	result := &Cache[T]{
		m:            make(map[string]storage[T], len(items)),
		ttl:          ttl,
		cleanupEvery: every,
		maxKeys:      maxKeys,
		stopCleaner:  make(chan struct{}),
	}

	now := time.Now()
	nowNano := now.UnixNano()
	needJanitor := ttl > 0
	for _, item := range items {
		if !f(item.v.s) {
			continue
		}
		cp := storage[T]{s: item.v.s}
		if override {
			cp.ttl = ttl
			if ttl > 0 {
				cp.expires = now.Add(ttl).UnixNano()
			}
		} else {
			cp.ttl = item.v.ttl
			cp.expires = item.v.expires
			if cp.ttl > 0 || cp.expires > 0 {
				needJanitor = true
			}
		}
		result.m[item.k] = cp
	}
	result.lastUpdated.Store(nowNano)
	if needJanitor {
		result.ensureCleaner()
	}
	return result
}

func (c *Cache[T]) FilterValues(f func(T) bool) []T {
	items, _ := c.snapshotLive()
	values := make([]T, 0, len(items))
	for _, item := range items {
		if f(item.v.s) {
			values = append(values, item.v.s)
		}
	}
	return values
}

func (c *Cache[T]) Dump(out io.Writer) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now().UnixNano()
	for k, v := range c.m {
		if isExpired(v, now) {
			continue
		}
		_, _ = fmt.Fprintf(out, "%s: %v\n", k, v.s)
	}
}

func (c *Cache[T]) cleanupExpired() {
	c.mu.RLock()
	now := time.Now().UnixNano()
	expiredKeys := make([]string, 0)
	for k, v := range c.m {
		if isExpired(v, now) {
			expiredKeys = append(expiredKeys, k)
		}
	}
	c.mu.RUnlock()

	if len(expiredKeys) == 0 {
		return
	}

	c.mu.Lock()
	deleted := false
	for _, k := range expiredKeys {
		if v, ok := c.m[k]; ok && isExpired(v, now) {
			delete(c.m, k)
			deleted = true
		}
	}
	if deleted {
		c.lastUpdated.Store(time.Now().UnixNano())
	}
	c.mu.Unlock()
}

func (c *Cache[T]) startCleaner(every time.Duration) {
	if every <= 0 {
		every = DefaultCleanupInterval
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
	if c == nil || c.stopCleaner == nil {
		return
	}
	c.stopOnce.Do(func() {
		close(c.stopCleaner)
	})
}
