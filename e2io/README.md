# e2io

讀取 Reader／檔案的 Must 輔助，以及目錄 fsnotify 監看（含 debounce）。

Must-style readers plus fsnotify directory watching with debounce.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2io
```

## 功能 / Features

- **讀取 / Read**：`MustReadAll`、`MustReadAllAsString`、`MustReadAllAndClose`、`MustReadFile`
- **監看 / Watch**：`WatchDir`（阻塞迴圈，於 goroutine 呼叫）

## 用法 / Usage

```go
import (
    "strings"
    "github.com/e2u/e2util/e2io"
)

b := e2io.MustReadAll(strings.NewReader("hello"))
s := e2io.MustReadFile("config.toml")
```
