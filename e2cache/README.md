# e2cache Documentation

## 項目概覽 / Project Overview

`e2cache` 是 `e2util` 工具庫中的一個子包，封裝了一個基於 `eko/gocache` 的緩存管理系統。它支持多種緩存後端，包括 Redis、內存緩存（基於 `patrickmn/go-cache`）和假緩存（`FakeCacheStore`），並通過統一的接口提供靈活的緩存操作。該包適用於需要高效鍵值存儲和訪問的應用場景，並允許根據配置動態選擇緩存類型。

`e2cache` is a sub-package of the `e2util` library, encapsulating a caching system based on `eko/gocache`. It supports multiple cache backends, including Redis, in-memory caching (via `patrickmn/go-cache`), and a fake cache (`FakeCacheStore`), offering a unified interface for flexible cache operations. This package is suitable for applications requiring efficient key-value storage and retrieval, with dynamic backend selection based on configuration.

---

## 使用方法 / Usage

### 1. 初始化緩存 / Initializing the Cache

Use the `New` function to create a cache instance based on configuration, supporting Redis, memory, or fake caches.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2cache"
)

func main() {
// Configure Redis cache
cfg := &e2cache.Config{
Enable: true,
Type:   "redis",
Dsn:    "redis://localhost:6379/0",
}
cacheConn := e2cache.New(cfg)
if cacheConn.Err != nil {
fmt.Println("Initialization failed:", cacheConn.Err)
return
}
fmt.Println("Redis cache initialized")
}
```

### 2. 設置和獲取緩存 / Setting and Getting Cache

Use `Set` and `Get` methods to store and retrieve cache data.

```go
package main

import (
"context"
"fmt"
"github.com/e2u/e2util/e2cache"
)

func main() {
cfg := &e2cache.Config{Type: "memory"}
cacheConn := e2cache.New(cfg)
ctx := context.Background()

// Set cache
err := cacheConn.Set(ctx, "user:1", "Alice", nil)
if err != nil {
fmt.Println("Set failed:", err)
return
}

// Get cache
value, err := cacheConn.Get(ctx, "user:1")
if err != nil {
fmt.Println("Get failed:", err)
return
}
fmt.Println("Cache value:", value) // "Alice"
}
```

### 3. 帶 TTL 的緩存操作 / Cache with TTL

Use `GetWithTTL` to retrieve a value and its remaining time-to-live, and set expiration via `store.Option`.

```go
package main

import (
"context"
"fmt"
"time"
"github.com/eko/gocache/lib/v4/store"
"github.com/e2u/e2util/e2cache"
)

func main() {
cfg := &e2cache.Config{Type: "memory"}
cacheConn := e2cache.New(cfg)
ctx := context.Background()

// Set cache with expiration
opts := []store.Option{store.WithExpiration(5 * time.Second)}
cacheConn.Set(ctx, "temp_key", 123, opts...)

// Get value and TTL
value, ttl, err := cacheConn.GetWithTTL(ctx, "temp_key")
if err != nil {
fmt.Println("Get failed:", err)
return
}
fmt.Println("Value:", value, "TTL:", ttl) // Value: 123 TTL: ~5s
}
```

### 4. 刪除和清空緩存 / Deleting and Clearing Cache

Use `Delete` to remove a single key and `Clear` to wipe all cache data.

```go
package main

import (
"context"
"fmt"
"github.com/e2u/e2util/e2cache"
)

func main() {
cfg := &e2cache.Config{Type: "memory"}
cacheConn := e2cache.New(cfg)
ctx := context.Background()

// Set and delete a single key
cacheConn.Set(ctx, "key1", "value1", nil)
cacheConn.Delete(ctx, "key1")
value, _ := cacheConn.Get(ctx, "key1")
fmt.Println("After delete:", value) // nil

// Set and clear all cache
cacheConn.Set(ctx, "key2", "value2", nil)
cacheConn.Clear(ctx)
value, _ = cacheConn.Get(ctx, "key2")
fmt.Println("After clear:", value) // nil
}
```

### 5. 使用假緩存 / Using Fake Cache

If an invalid cache type is specified, it defaults to `FakeCacheStore`, simulating cache behavior without storing data.

```go
package main

import (
"context"
"fmt"
"github.com/e2u/e2util/e2cache"
)

func main() {
cfg := &e2cache.Config{Type: "unknown"}
cacheConn := e2cache.New(cfg)
ctx := context.Background()

// Set and get using fake cache
cacheConn.Set(ctx, "key", "value", nil)
value, _ := cacheConn.Get(ctx, "key")
fmt.Println("Fake cache value:", value) // nil
}
```

---

## 安裝步驟 / Installation Steps

1. **確保 Go 環境**
確認已安裝 Go（建議使用 1.16 或以上版本），並設置好 `GOPATH`。
2. **下載項目**
在終端運行以下命令：
```bash
go get -u github.com/e2u/e2util/e2cache
```
3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2cache`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。

1. **Ensure Go Environment**
Confirm Go (version 1.16 or higher recommended) is installed and `GOPATH` is set.
2. **Download the Package**
Run the following command in your terminal:
```bash
go get -u github.com/e2u/e2util/e2cache
```
3. **Verify Installation**
Import `github.com/e2u/e2util/e2cache` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **多後端支持**：支援 Redis（`redis`）、內存（`memory`）和假緩存（`fake`）三種存儲類型。
- **緩存設置**：`Set` 方法存儲鍵值對，支持過期時間配置。
- **緩存獲取**：`Get` 和 `GetWithTTL` 方法檢索緩存值及其剩餘有效時間。
- **緩存管理**：提供 `Delete`（刪除單鍵）、`Invalidate`（失效緩存）和 `Clear`（清空緩存）功能。
- **假緩存**：`FakeCacheStore` 提供無實際存儲的模擬實現，方便測試或禁用緩存。

- **Multi-Backend Support**: Supports Redis (`redis`), in-memory (`memory`), and fake (`fake`) storage types.
- **Cache Setting**: The `Set` method stores key-value pairs with optional expiration.
- **Cache Retrieval**: The `Get` and `GetWithTTL` methods retrieve values and their remaining TTL.
- **Cache Management**: Offers `Delete` (remove single key), `Invalidate` (invalidate cache), and `Clear` (wipe all cache) operations.
- **Fake Cache**: `FakeCacheStore` provides a no-op implementation for testing or disabling caching.

---
```
