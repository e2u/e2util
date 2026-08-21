# e2gocache

進程內泛型快取（自己維護 `map` + TTL）。唔係 `e2cache`：`e2cache` 係 Redis／memory／fake 嘅應用層封裝，畀 `e2app` 用。

In-process generic cache (own `map` + TTL). This is not `e2cache`, which wraps Redis / memory / fake for `e2app`.

需要 Go 1.27+（`Cache.GetAS[T]` 用 generic method）。

Requires Go 1.27+ (`Cache.GetAS[T]` is a generic method).

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2gocache
```

## 功能 / Features

- 泛型 `Cache[T]`：`Set` / `Get` / `Del` / `GetOrSet` / `Flush`
- Generic `Cache[T]`: `Set` / `Get` / `Del` / `GetOrSet` / `Flush`
- 預設 TTL、單 key 覆寫 TTL、永不過期
- Default TTL, per-key override, never-expire
- Lazy delete（`Get` 遇到過期即刪）+ 可選背景 janitor
- Lazy delete on `Get`, plus an optional janitor
- `Filter` / `FilterValues`：snapshot 後解鎖再跑 predicate
- `Filter` / `FilterValues`: snapshot live entries, unlock, then run the predicate
- `WithMaxKeys` 上限；超出時先丟過期再淘汰一條
- `WithMaxKeys` cap; over the cap, expired keys go first, then one arbitrary key
- `GetAS[T]` 對 `Cache[any]`（或任何 `T`）做類型斷言
- `GetAS[T]` type-asserts a stored value

## 快速開始 / Quick start

```go
package main

import (
	"fmt"
	"time"

	"github.com/e2u/e2util/e2gocache"
)

func main() {
	c := e2gocache.New[string](time.Minute)
	defer c.Stop() // 停背景 janitor / stop the janitor

	c.Set("k", "v")
	c.Set("short", "x", 5*time.Second) // 覆寫 TTL / per-key TTL
	v, ok := c.Get("k")
	fmt.Println(v, ok)

	v = c.GetOrSet("computed", func() string { return "from-fn" })
	c.WithMaxKeys(10_000)
	c.Del("k")

	// mixed types: GetAS type-asserts the stored value
	mixed := e2gocache.New[any](time.Minute)
	defer mixed.Stop()
	mixed.Set("n", 42)
	n, ok := mixed.GetAS[int]("n")
	fmt.Println(n, ok)
}
```

## 建構與 janitor / Construction and janitor

```go
c := e2gocache.New[string](time.Minute)                      // 預設 TTL 1 分鐘，即開 janitor
c := e2gocache.New[string](0)                                // 永不過期，lazy 開 janitor
c := e2gocache.New[string](time.Minute, 100*time.Millisecond) // 自訂清理週期
```

| `New` 參數 / argument | 語意 / meaning |
| --- | --- |
| `expiration > 0` | 預設每條 key 嘅 TTL；janitor **即開** / default per-key TTL; janitor starts immediately |
| `expiration <= 0` | 預設永不過期（`DefaultExpiration`）；janitor **唔開**，直到有 key 帶正 TTL / never expire; janitor waits for a positive TTL |
| 第二個 `cleanup` | janitor 週期；`<= 0` 當冇傳 / janitor period; `<= 0` is ignored |

自動週期：預設 1 秒（`DefaultCleanupInterval`）。若預設 TTL 短過 1 秒，週期跟 TTL，最短 10ms。

The janitor defaults to 1s. If the default TTL is shorter than 1s, cleanup follows that TTL (minimum 10ms).

`Set` / `GetOrSet` 第三個參數 `<= 0` 表示呢條 key 永不過期，會蓋過 `New` 嘅預設。

A `Set` / `GetOrSet` TTL of `<= 0` means this key never expires (overrides the cache default).

用完請 `Stop()`，避免 janitor goroutine 洩漏。對未開 janitor 或者 `nil` cache 呼叫都安全，重複 `Stop()` 亦唔會 panic。

Call `Stop()` when done so the janitor goroutine does not leak. `Stop` is safe on `nil`, on a cache that never started a janitor, and when called more than once.

## API

| 方法 / Method | 說明 / Description |
| --- | --- |
| `New[T](expiration, cleanup...)` | 建立 cache |
| `Stop()` | 停 janitor |
| `Set(key, value, expiration...)` | 寫入；可選 TTL 覆寫 |
| `Get(key) (T, bool)` | 讀取；過期當 miss 並刪除 |
| `GetOrSet(key, fn, expiration...) T` | miss 時算一次再寫入 |
| `GetAS[Out](key) (Out, bool)` | 類型斷言取出 |
| `GetAS[Out](c, key)` | 套件級別名 / package alias |
| `Del(key)` | 刪除 |
| `Len() int` | 未過期條目數 |
| `Keys() []string` | 未過期 keys |
| `Values() []T` | 未過期 values |
| `Filter(f, expiration...) *Cache[T]` | 複製符合條件嘅項到新 cache |
| `FilterValues(f) []T` | 只回符合條件嘅 values |
| `WithMaxKeys(n) *Cache[T]` | 上限；`<= 0` 不限 |
| `Flush()` | 清空 |
| `Dump(w)` | 把未過期項寫到 `io.Writer` |
| `LastUpdated() time.Time` | 最近一次寫入／刪除時間 |
| `LastUpdatedString() string` | 同上，UTC RFC3339Nano |

常數 / constants：`DefaultExpiration`（`-1`，永不過期）、`DefaultCleanupInterval`（`1s`）。

### GetOrSet

miss 時喺**寫鎖內**呼叫 `fn`，並發同一 key 只算一次。`fn` **唔好**再入同一 cache（會死鎖）。hit 時唔跑 `fn`。過期 key 當 miss，會重新計算。

On miss, `fn` runs under the **write lock** (one compute per key). Do **not** call back into the same cache from `fn` (deadlock). Hits skip `fn`. An expired key is a miss and is recomputed.

### GetAS

`Cache.GetAS[T]` 用 Go 1.27 generic method。`T` 係你要取出嘅類型；key 唔存在、已過期、或類型唔啱時 `ok == false`。套件級 `GetAS[T](c, key)` 仍然可用；`c == nil` 當 miss。

`Cache.GetAS[T]` is a Go 1.27 generic method. `T` is the desired type; `ok` is false on a miss, expiry, or failed assertion. The package function `GetAS[T](c, key)` remains as an alias. A nil cache is a miss.

```go
mixed := e2gocache.New[any](time.Minute)
defer mixed.Stop()
mixed.Set("n", 42)
n, ok := mixed.GetAS[int]("n")      // 42, true
_, ok = mixed.GetAS[string]("n")    // false
_, ok = e2gocache.GetAS[int](mixed, "n")
```

### Filter / FilterValues

```go
odds := c.Filter(func(n int) bool { return n%2 == 1 })
defer odds.Stop()

vals := c.FilterValues(func(s string) bool { return strings.HasPrefix(s, "a") })

// 覆寫結果 cache 嘅 TTL / override TTL on the result
short := c.Filter(func(string) bool { return true }, 10*time.Second)
defer short.Stop()
```

- 只複製**未過期**項。 / Only live entries are copied.
- 先 snapshot 再解鎖，然後先跑 `f`。`f` 可以再 `Get`／`Set` 同一 cache，唔會因為佔住讀鎖而死鎖。 / Snapshot, unlock, then run `f`. The predicate may re-enter the same cache.
- 無 TTL 參數時，結果保留每條 key 嘅**剩餘**過期時間（絕對 `expires`），並複製 `maxKeys`。 / With no TTL argument, remaining expiry is preserved and `maxKeys` is copied.
- 有 TTL 參數時，結果用新 TTL（`<= 0` 表示結果內各項永不過期）。 / A TTL argument resets expiry on the result (`<= 0` means never).
- 回傳嘅 `*Cache[T]` 係獨立實例，之後改來源唔影響結果。用完要 `Stop()`。 / The result is independent of the source. Call `Stop()` on it.
- 來源同結果都係永不過期時，結果**唔會**開 janitor。 / A never-expire result does not start a janitor.

### WithMaxKeys

```go
c := e2gocache.New[int](0).WithMaxKeys(10_000)
```

`n <= 0` 表示不限。插入**新 key** 且已滿時：先刪過期，再隨機刪一條。更新已有 key 唔會觸發淘汰。對已有資料呼叫 `WithMaxKeys(n)` 會即刻收縮。

`n <= 0` means unlimited. Inserting a **new** key at the cap drops expired keys first, then one arbitrary key. Updating an existing key does not evict. Calling `WithMaxKeys(n)` on an already-full cache shrinks immediately.

淘汰係任意一條（map 迭代順序），唔係 LRU。

Eviction is arbitrary (map iteration order), not LRU.

## 過期語意 / Expiry

- `Get` 遇到過期 key 會 **lazy delete**，回 miss。
- `Len` / `Keys` / `Values` / `Filter` / `FilterValues` / `Dump` 都**唔計**過期項；佢哋唔會順便刪（除咗 `Get`）。
- `Len` / `Keys` / `Values` / `Filter` / `FilterValues` / `Dump` skip expired entries; they do not delete them (`Get` does).
- 背景 janitor 按週期再掃，真正從 map 移除過期 key。
- The janitor periodically removes expired keys from the map.
- `LastUpdated` 喺 `Set`、`GetOrSet` 寫入、`Del`、`Flush`、以及 lazy／janitor 刪除時更新。`Get` hit 唔改。
- `LastUpdated` moves on writes, `Del`, `Flush`, and expiry deletes. A `Get` hit does not change it.

## 執行緒安全 / Thread safety

所有公開方法都係 goroutine-safe。`Cache[T]` 用 `RWMutex`。

All exported methods are goroutine-safe. `Cache[T]` uses an `RWMutex`.

| 回調 / callback | 鎖 / lock | 可唔可以再入同一 cache / re-enter same cache |
| --- | --- | --- |
| `GetOrSet` 嘅 `fn` | 寫鎖 / write lock | 唔得，會死鎖 / no, deadlock |
| `Filter` / `FilterValues` 嘅 `f` | 無鎖（snapshot 之後） / none after snapshot | 得 / yes |

`T` 按賦值複製。若 `T` 係 pointer、slice、map，來源同 cache 會共用底層資料。

`T` is copied by assignment. If `T` is a pointer, slice, or map, the caller and the cache share the underlying data.

## 同 e2cache 嘅分別 / vs e2cache

| | `e2gocache` | `e2cache` |
| --- | --- | --- |
| 範圍 / Scope | 單進程 / in-process | Redis 或 memory store |
| 類型 / Type | 泛型 `T` | `any` + eko/gocache |
| 用家 / Used by | 熱數據、本地計算 / hot local data | `e2app` 設定 `[cache]` |
| 過期 / Expiry | 自己嘅 TTL + janitor | store 實作 |
| 分佈式 / Distributed | 唔係 / no | Redis 可以 / Redis can |

本地熱數據、請求內去重、計算結果用 `e2gocache`。要跨進程或者跟 `e2app` 設定走，用 `e2cache`。

Use `e2gocache` for in-process hot data, request coalescing, and memoization. Use `e2cache` when you need a shared store or `e2app` `[cache]` config.
