# e2sync

`sync.Map` 輔助函數，以及泛型紅黑樹 map。

Helpers for `sync.Map` plus a generic red-black tree map.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2sync
```

## 功能 / Features

- **sync.Map**：`SyncMapLen`、`SyncMapSortStringKeys`、原子 `Add`
- **RBTreeMap**：有序鍵值儲存，需提供 `Comparer`

## 用法 / Usage

```go
import (
    "sync"
    "github.com/e2u/e2util/e2sync"
)

var m sync.Map
e2sync.Add(&m, "hits", int64(1))
n := e2sync.SyncMapLen(&sync.RWMutex{}, &m)
```
