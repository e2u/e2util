# e2logrus

以設定檔建立 logrus Logger：JSON／text、輪轉檔、序號 Hook。

Build a logrus Logger from config: JSON/text, rotating files, sequence hook.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2logrus
```

## 功能 / Features

- **NewLogger**：`Output` 可為 `stdout` 或 `file:///path/log.%Y%m%d`
- **CloneLogrus**：複製 formatter／level／output
- **SeqHook**：每條日誌加 `seq` 欄位

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2logrus"

log := e2logrus.NewLogger(&e2logrus.Config{
    Output: "stdout",
    Level:  "info",
    Format: "json",
})
log.Info("started")
```
