package e2gocache

import (
	"bytes"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetAS(t *testing.T) {
	c := New[any](0)
	defer c.Stop()
	c.Set("s", "hello")
	c.Set("n", 42)

	s, ok := c.GetAS[string]("s")
	if !ok || s != "hello" {
		t.Fatalf("GetAS[string] = %q, %v", s, ok)
	}
	n, ok := c.GetAS[int]("n")
	if !ok || n != 42 {
		t.Fatalf("GetAS[int] = %d, %v", n, ok)
	}
	if _, ok := c.GetAS[int]("s"); ok {
		t.Fatal("wrong type should miss")
	}
	if _, ok := c.GetAS[string]("missing"); ok {
		t.Fatal("missing key should miss")
	}

	typed := New[string](0)
	defer typed.Stop()
	typed.Set("k", "v")
	got, ok := typed.GetAS[string]("k")
	if !ok || got != "v" {
		t.Fatalf("GetAS on Cache[string] = %q, %v", got, ok)
	}
	if _, ok := typed.GetAS[int]("k"); ok {
		t.Fatal("Cache[string] as int should miss")
	}

	var nilCache *Cache[any]
	if _, ok := nilCache.GetAS[string]("k"); ok {
		t.Fatal("nil cache should miss")
	}
	if _, ok := GetAS[int](c, "n"); !ok {
		t.Fatal("package GetAS should still work")
	}
}

func TestCleanupInterval(t *testing.T) {
	c := New[int](time.Minute)
	defer c.Stop()
	if c.cleanupEvery != time.Second {
		t.Fatalf("default TTL 1m -> cleanup %v, want 1s", c.cleanupEvery)
	}

	short := New[int](50 * time.Millisecond)
	defer short.Stop()
	if short.cleanupEvery != 50*time.Millisecond {
		t.Fatalf("TTL 50ms -> cleanup %v", short.cleanupEvery)
	}

	tiny := New[int](time.Nanosecond)
	defer tiny.Stop()
	if tiny.cleanupEvery != minCleanupInterval {
		t.Fatalf("tiny TTL -> cleanup %v, want %v", tiny.cleanupEvery, minCleanupInterval)
	}

	custom := New[int](time.Minute, 5*time.Second)
	defer custom.Stop()
	if custom.cleanupEvery != 5*time.Second {
		t.Fatalf("explicit cleanup %v", custom.cleanupEvery)
	}
}

func TestSetGetDel(t *testing.T) {
	c := New[string](0)
	defer c.Stop()

	c.Set("k", "v")
	got, ok := c.Get("k")
	if !ok || got != "v" {
		t.Fatalf("Get = %q, %v", got, ok)
	}
	if c.Len() != 1 {
		t.Fatalf("Len = %d", c.Len())
	}
	c.Del("k")
	if _, ok := c.Get("k"); ok {
		t.Fatal("expected miss after Del")
	}
	if c.Len() != 0 {
		t.Fatalf("Len after Del = %d", c.Len())
	}
}

func TestGetDeletesExpired(t *testing.T) {
	c := New[int](0)
	defer c.Stop()

	c.Set("k", 1, 20*time.Millisecond)
	if _, ok := c.Get("k"); !ok {
		t.Fatal("expected hit before expiry")
	}
	time.Sleep(40 * time.Millisecond)
	if _, ok := c.Get("k"); ok {
		t.Fatal("expected miss after expiry")
	}
	c.mu.RLock()
	_, stillThere := c.m["k"]
	c.mu.RUnlock()
	if stillThere {
		t.Fatal("expired key should be deleted on Get")
	}
}

func TestLenKeysValuesSkipExpired(t *testing.T) {
	c := New[string](0)
	defer c.Stop()

	c.Set("live", "a", time.Hour)
	c.Set("dead", "b", 10*time.Millisecond)
	time.Sleep(25 * time.Millisecond)

	if c.Len() != 1 {
		t.Fatalf("Len = %d, want 1", c.Len())
	}
	keys := c.Keys()
	if len(keys) != 1 || keys[0] != "live" {
		t.Fatalf("Keys = %v", keys)
	}
	vals := c.Values()
	if len(vals) != 1 || vals[0] != "a" {
		t.Fatalf("Values = %v", vals)
	}
	filtered := c.FilterValues(func(s string) bool { return true })
	if len(filtered) != 1 || filtered[0] != "a" {
		t.Fatalf("FilterValues = %v", filtered)
	}
}

func TestFilterAndStop(t *testing.T) {
	c := New[int](time.Hour)
	defer c.Stop()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	f := c.Filter(func(n int) bool { return n%2 == 1 })
	defer f.Stop()

	if f.Len() != 2 {
		t.Fatalf("Filter Len = %d", f.Len())
	}
	if _, ok := f.Get("a"); !ok {
		t.Fatal("expected a")
	}
	if _, ok := f.Get("b"); ok {
		t.Fatal("b should be filtered out")
	}
	f.Stop() // second Stop must not panic
}

func TestFilterCallbackDoesNotHoldLock(t *testing.T) {
	c := New[int](time.Hour)
	defer c.Stop()
	c.Set("a", 1)
	c.Set("b", 2)

	done := make(chan struct{})
	go func() {
		defer close(done)
		f := c.Filter(func(n int) bool {
			c.Set("from-filter", n)
			_, _ = c.Get("a")
			return n == 1
		})
		defer f.Stop()
		if f.Len() != 1 {
			t.Errorf("Filter Len = %d", f.Len())
		}
		vals := c.FilterValues(func(n int) bool {
			_, _ = c.Get("b")
			c.Del("from-filter")
			return n > 0
		})
		if len(vals) == 0 {
			t.Error("FilterValues returned empty")
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Filter/FilterValues callback deadlocked on the same cache")
	}
}

func TestFilterPreservesRemainingTTL(t *testing.T) {
	c := New[string](0)
	defer c.Stop()
	c.Set("k", "v", 80*time.Millisecond)

	f := c.Filter(func(string) bool { return true })
	defer f.Stop()

	time.Sleep(30 * time.Millisecond)
	if _, ok := f.Get("k"); !ok {
		t.Fatal("should still be alive")
	}
	time.Sleep(70 * time.Millisecond)
	if _, ok := f.Get("k"); ok {
		t.Fatal("should expire with original remaining TTL")
	}
}

func TestDumpPrintsValue(t *testing.T) {
	c := New[string](0)
	defer c.Stop()
	c.Set("hello", "world")
	c.Set("gone", "x", time.Nanosecond)
	time.Sleep(5 * time.Millisecond)

	var buf bytes.Buffer
	c.Dump(&buf)
	s := buf.String()
	if !strings.Contains(s, "hello: world") {
		t.Fatalf("Dump = %q", s)
	}
	if strings.Contains(s, "gone") || strings.Contains(s, "storage") {
		t.Fatalf("Dump should skip expired and not print storage: %q", s)
	}
}

func TestFlushAndLastUpdated(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.Set("a", 1)
	before := c.LastUpdated()
	time.Sleep(2 * time.Millisecond)
	c.Flush()
	if c.Len() != 0 {
		t.Fatal("Flush should empty cache")
	}
	if !c.LastUpdated().After(before) {
		t.Fatalf("LastUpdated did not move: %v -> %v", before, c.LastUpdated())
	}
	_ = c.LastUpdatedString()
}

func TestConcurrentSetGet(t *testing.T) {
	c := New[int](time.Minute)
	defer c.Stop()

	var wg sync.WaitGroup
	for i := range 32 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := range 50 {
				k := "k"
				c.Set(k, i+j)
				_, _ = c.Get(k)
			}
		}(i)
	}
	wg.Wait()
	if c.Len() != 1 {
		t.Fatalf("Len = %d", c.Len())
	}
}

func TestStopNilSafe(t *testing.T) {
	var c *Cache[int]
	c.Stop()
	(&Cache[int]{}).Stop()
}

func TestLazyJanitor(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	if c.janitorOn.Load() {
		t.Fatal("New(0) should not start janitor")
	}
	c.Set("a", 1)
	if c.janitorOn.Load() {
		t.Fatal("Set without TTL should not start janitor")
	}
	c.Set("b", 2, time.Hour)
	if !c.janitorOn.Load() {
		t.Fatal("Set with TTL should start janitor")
	}

	expiring := New[int](time.Minute)
	defer expiring.Stop()
	if !expiring.janitorOn.Load() {
		t.Fatal("New with default TTL should start janitor")
	}
}

func TestGetOrSet(t *testing.T) {
	c := New[int](0)
	defer c.Stop()

	if v := c.GetOrSet("k", func() int { return 7 }); v != 7 {
		t.Fatalf("GetOrSet miss = %d", v)
	}
	if v := c.GetOrSet("k", func() int { t.Fatal("fn should not run"); return 0 }); v != 7 {
		t.Fatalf("GetOrSet hit = %d", v)
	}

	var calls atomic.Int32
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			c.GetOrSet("once", func() int {
				calls.Add(1)
				time.Sleep(5 * time.Millisecond)
				return 9
			})
		})
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("concurrent GetOrSet fn calls = %d, want 1", calls.Load())
	}
}

func TestMaxKeys(t *testing.T) {
	c := New[int](0).WithMaxKeys(2)
	defer c.Stop()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	if c.Len() != 2 {
		t.Fatalf("Len = %d, want 2", c.Len())
	}
	c.Set("a", 10) // update existing, still 2
	if c.Len() != 2 {
		t.Fatalf("Len after update = %d", c.Len())
	}
	got, ok := c.Get("a")
	if !ok || got != 10 {
		t.Fatalf("updated a = %d, %v", got, ok)
	}
}

func eventually(t *testing.T, d time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(d)
	for {
		if fn() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("condition not met before timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestConstants(t *testing.T) {
	if DefaultExpiration != -1 {
		t.Fatalf("DefaultExpiration = %v, want -1", DefaultExpiration)
	}
	if DefaultCleanupInterval != time.Second {
		t.Fatalf("DefaultCleanupInterval = %v, want 1s", DefaultCleanupInterval)
	}
}

func TestCleanupIntervalHelper(t *testing.T) {
	tests := []struct {
		name       string
		expiration time.Duration
		override   []time.Duration
		want       time.Duration
	}{
		{"default TTL 1m", time.Minute, nil, DefaultCleanupInterval},
		{"TTL 50ms follows TTL", 50 * time.Millisecond, nil, 50 * time.Millisecond},
		{"tiny TTL floors at 10ms", time.Nanosecond, nil, minCleanupInterval},
		{"exactly 10ms", minCleanupInterval, nil, minCleanupInterval},
		{"exactly 1s", time.Second, nil, DefaultCleanupInterval},
		{"no default TTL", 0, nil, DefaultCleanupInterval},
		{"never-expire sentinel", DefaultExpiration, nil, DefaultCleanupInterval},
		{"explicit override", time.Minute, []time.Duration{5 * time.Second}, 5 * time.Second},
		{"zero override ignored", time.Minute, []time.Duration{0}, DefaultCleanupInterval},
		{"negative override ignored", time.Minute, []time.Duration{-1}, DefaultCleanupInterval},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanupInterval(tt.expiration, tt.override...)
			if got != tt.want {
				t.Fatalf("cleanupInterval(%v, %v) = %v, want %v", tt.expiration, tt.override, got, tt.want)
			}
		})
	}
}

func TestNewNeverExpire(t *testing.T) {
	for _, exp := range []time.Duration{0, DefaultExpiration, -time.Hour} {
		c := New[int](exp)
		defer c.Stop()
		if c.ttl != DefaultExpiration {
			t.Fatalf("New(%v) ttl = %v, want %v", exp, c.ttl, DefaultExpiration)
		}
		if c.janitorOn.Load() {
			t.Fatalf("New(%v) should not start janitor", exp)
		}
		if c.cleanupEvery != DefaultCleanupInterval {
			t.Fatalf("New(%v) cleanupEvery = %v", exp, c.cleanupEvery)
		}
	}
}

func TestNewExplicitCleanupZeroIgnored(t *testing.T) {
	c := New[int](time.Minute, 0)
	defer c.Stop()
	if c.cleanupEvery != DefaultCleanupInterval {
		t.Fatalf("cleanupEvery = %v, want %v", c.cleanupEvery, DefaultCleanupInterval)
	}
}

func TestGetMiss(t *testing.T) {
	c := New[string](0)
	defer c.Stop()
	got, ok := c.Get("missing")
	if ok || got != "" {
		t.Fatalf("Get missing = %q, %v", got, ok)
	}
}

func TestSetUsesDefaultTTL(t *testing.T) {
	c := New[int](40 * time.Millisecond)
	defer c.Stop()
	c.Set("k", 1)
	if _, ok := c.Get("k"); !ok {
		t.Fatal("expected hit before default TTL")
	}
	eventually(t, time.Second, func() bool {
		_, ok := c.Get("k")
		return !ok
	})
}

func TestSetZeroTTLOverridesDefault(t *testing.T) {
	c := New[int](20 * time.Millisecond)
	defer c.Stop()
	c.Set("k", 1, 0)
	time.Sleep(50 * time.Millisecond)
	got, ok := c.Get("k")
	if !ok || got != 1 {
		t.Fatalf("Set(..., 0) should never expire, got %d, %v", got, ok)
	}
}

func TestSetOverwriteTTL(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.Set("k", 1, 20*time.Millisecond)
	c.Set("k", 2, time.Hour)
	time.Sleep(40 * time.Millisecond)
	got, ok := c.Get("k")
	if !ok || got != 2 {
		t.Fatalf("overwrite with long TTL = %d, %v", got, ok)
	}
}

func TestKeysValuesMultiple(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.Set("b", 2)
	c.Set("a", 1)

	keys := c.Keys()
	sort.Strings(keys)
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("Keys = %v", keys)
	}
	vals := c.Values()
	sort.Ints(vals)
	if len(vals) != 2 || vals[0] != 1 || vals[1] != 2 {
		t.Fatalf("Values = %v", vals)
	}
}

func TestEmptyCacheViews(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	if c.Len() != 0 {
		t.Fatalf("Len = %d", c.Len())
	}
	if keys := c.Keys(); len(keys) != 0 {
		t.Fatalf("Keys = %v", keys)
	}
	if vals := c.Values(); len(vals) != 0 {
		t.Fatalf("Values = %v", vals)
	}
	if vals := c.FilterValues(func(int) bool { return true }); len(vals) != 0 {
		t.Fatalf("FilterValues = %v", vals)
	}
	f := c.Filter(func(int) bool { return true })
	defer f.Stop()
	if f.Len() != 0 {
		t.Fatalf("Filter Len = %d", f.Len())
	}
	var buf bytes.Buffer
	c.Dump(&buf)
	if buf.Len() != 0 {
		t.Fatalf("Dump empty = %q", buf.String())
	}
}

func TestGetOrSetExpirationAndRecompute(t *testing.T) {
	c := New[int](0)
	defer c.Stop()

	if v := c.GetOrSet("k", func() int { return 1 }, 30*time.Millisecond); v != 1 {
		t.Fatalf("GetOrSet = %d", v)
	}
	if !c.janitorOn.Load() {
		t.Fatal("GetOrSet with TTL should start janitor")
	}
	eventually(t, time.Second, func() bool {
		_, ok := c.Get("k")
		return !ok
	})
	if v := c.GetOrSet("k", func() int { return 2 }); v != 2 {
		t.Fatalf("GetOrSet after expiry = %d, want 2", v)
	}
}

func TestGetOrSetAfterDel(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.GetOrSet("k", func() int { return 1 })
	c.Del("k")
	if v := c.GetOrSet("k", func() int { return 2 }); v != 2 {
		t.Fatalf("GetOrSet after Del = %d", v)
	}
}

func TestGetASExpired(t *testing.T) {
	c := New[any](0)
	defer c.Stop()
	c.Set("n", 42, 20*time.Millisecond)
	eventually(t, time.Second, func() bool {
		_, ok := c.GetAS[int]("n")
		return !ok
	})
	var nilCache *Cache[any]
	if _, ok := GetAS[int](nilCache, "n"); ok {
		t.Fatal("package GetAS on nil cache should miss")
	}
}

func TestFilterSkipsExpired(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.Set("live", 1, time.Hour)
	c.Set("dead", 2, 15*time.Millisecond)
	eventually(t, time.Second, func() bool {
		return c.Len() == 1
	})
	f := c.Filter(func(int) bool { return true })
	defer f.Stop()
	if f.Len() != 1 {
		t.Fatalf("Filter Len = %d, want 1", f.Len())
	}
	if _, ok := f.Get("dead"); ok {
		t.Fatal("expired key should not be copied")
	}
}

func TestFilterValuesPredicate(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	vals := c.FilterValues(func(n int) bool { return n >= 2 })
	sort.Ints(vals)
	if len(vals) != 2 || vals[0] != 2 || vals[1] != 3 {
		t.Fatalf("FilterValues = %v", vals)
	}
	none := c.FilterValues(func(int) bool { return false })
	if len(none) != 0 {
		t.Fatalf("no-match FilterValues = %v", none)
	}
}

func TestFilterExpirationOverride(t *testing.T) {
	c := New[string](0)
	defer c.Stop()
	c.Set("k", "v")

	f := c.Filter(func(string) bool { return true }, 40*time.Millisecond)
	defer f.Stop()
	if !f.janitorOn.Load() {
		t.Fatal("positive override TTL should start janitor")
	}
	if _, ok := f.Get("k"); !ok {
		t.Fatal("expected live after Filter override")
	}
	eventually(t, time.Second, func() bool {
		_, ok := f.Get("k")
		return !ok
	})
}

func TestFilterOverrideNeverExpire(t *testing.T) {
	c := New[string](0)
	defer c.Stop()
	c.Set("k", "v", 20*time.Millisecond)

	f := c.Filter(func(string) bool { return true }, 0)
	defer f.Stop()
	time.Sleep(50 * time.Millisecond)
	got, ok := f.Get("k")
	if !ok || got != "v" {
		t.Fatalf("override 0 should never expire, got %q, %v", got, ok)
	}
	if _, ok := c.Get("k"); ok {
		t.Fatal("source key should have expired")
	}
}

func TestFilterNeverExpireNoJanitor(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.Set("a", 1)
	f := c.Filter(func(int) bool { return true })
	defer f.Stop()
	if f.janitorOn.Load() {
		t.Fatal("never-expire Filter result should not start janitor")
	}
}

func TestFilterCopiesMaxKeys(t *testing.T) {
	c := New[int](0).WithMaxKeys(2)
	defer c.Stop()
	c.Set("a", 1)
	c.Set("b", 2)

	f := c.Filter(func(int) bool { return true })
	defer f.Stop()
	f.Set("c", 3)
	if f.Len() != 2 {
		t.Fatalf("Filter should copy maxKeys=2, Len = %d", f.Len())
	}
}

func TestFilterResultIndependent(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.Set("a", 1)
	f := c.Filter(func(int) bool { return true })
	defer f.Stop()
	c.Set("a", 99)
	c.Del("a")
	got, ok := f.Get("a")
	if !ok || got != 1 {
		t.Fatalf("Filter snapshot should keep original value, got %d, %v", got, ok)
	}
}

func TestMaxKeysEvictsExpiredFirst(t *testing.T) {
	c := New[int](0).WithMaxKeys(2)
	defer c.Stop()
	c.Set("dead", 1, 15*time.Millisecond)
	c.Set("live", 2, time.Hour)
	eventually(t, time.Second, func() bool {
		return c.Len() == 1
	})
	c.mu.RLock()
	_, deadStill := c.m["dead"]
	c.mu.RUnlock()
	if !deadStill {
		t.Fatal("expired key should still be in the map before insert")
	}
	c.Set("new", 3)
	if _, ok := c.Get("live"); !ok {
		t.Fatal("live key should remain")
	}
	if _, ok := c.Get("new"); !ok {
		t.Fatal("new key should be stored after dropping expired")
	}
	if _, ok := c.Get("dead"); ok {
		t.Fatal("expired key should have been evicted")
	}
	if c.Len() != 2 {
		t.Fatalf("Len = %d, want 2", c.Len())
	}
}

func TestWithMaxKeysShrinksExisting(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.WithMaxKeys(1)
	if c.Len() != 1 {
		t.Fatalf("Len after shrink = %d, want 1", c.Len())
	}
}

func TestWithMaxKeysExactCapKeepsAll(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.Set("a", 1)
	c.Set("b", 2)
	c.WithMaxKeys(2)
	if c.Len() != 2 {
		t.Fatalf("Len at exact cap = %d, want 2", c.Len())
	}
}

func TestWithMaxKeysUnlimited(t *testing.T) {
	c := New[int](0).WithMaxKeys(1)
	defer c.Stop()
	c.Set("a", 1)
	c.Set("b", 2)
	if c.Len() != 1 {
		t.Fatalf("Len = %d, want 1", c.Len())
	}
	c.WithMaxKeys(0)
	c.Set("c", 3)
	c.Set("d", 4)
	if c.Len() != 3 {
		t.Fatalf("unlimited after WithMaxKeys(0), Len = %d", c.Len())
	}
}

func TestWithMaxKeysNil(t *testing.T) {
	var c *Cache[int]
	if c.WithMaxKeys(2) != nil {
		t.Fatal("nil WithMaxKeys should return nil")
	}
}

func TestJanitorRemovesExpired(t *testing.T) {
	c := New[int](30*time.Millisecond, 15*time.Millisecond)
	defer c.Stop()
	c.Set("k", 1)
	eventually(t, time.Second, func() bool {
		c.mu.RLock()
		n := len(c.m)
		c.mu.RUnlock()
		return n == 0
	})
}

func TestCleanupExpiredSkipsLive(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.Set("k", 1, time.Nanosecond)
	time.Sleep(5 * time.Millisecond)
	c.Set("k", 2, time.Hour)
	c.cleanupExpired()
	got, ok := c.Get("k")
	if !ok || got != 2 {
		t.Fatalf("refreshed key deleted by cleanup, got %d, %v", got, ok)
	}
}

func TestLastUpdatedStringFormat(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	if c.LastUpdated().UnixNano() != 0 {
		t.Fatalf("LastUpdated before Set = %v", c.LastUpdated())
	}
	c.Set("a", 1)
	s := c.LastUpdatedString()
	parsed, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t.Fatalf("LastUpdatedString = %q: %v", s, err)
	}
	if parsed.Location() != time.UTC {
		t.Fatalf("LastUpdatedString location = %v", parsed.Location())
	}
}

func TestDelMissingUpdatesLastUpdated(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.Set("a", 1)
	before := c.LastUpdated()
	time.Sleep(2 * time.Millisecond)
	c.Del("missing")
	if !c.LastUpdated().After(before) {
		t.Fatal("Del should bump LastUpdated even if key is missing")
	}
}

func TestStopNeverStartedJanitor(t *testing.T) {
	c := New[int](0)
	c.Stop()
	c.Stop()
	c.Set("a", 1)
}

func TestDumpInteger(t *testing.T) {
	c := New[int](0)
	defer c.Stop()
	c.Set("n", 42)
	var buf bytes.Buffer
	c.Dump(&buf)
	if !strings.Contains(buf.String(), "n: 42") {
		t.Fatalf("Dump = %q", buf.String())
	}
}
