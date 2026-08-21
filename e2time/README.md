# e2time

日期解析、當日零點、加減日、時間指標、隨機 sleep。

Date parse, start of today, add days, time pointer, random sleep.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2time
```

## 功能 / Features

- `MustParse`：失敗回傳 zero time
- `ToDay`、`AddDay`
- `TimePointer`
- `SleepRandom(min, max)`（含兩端）

## 用法 / Usage

```go
import (
    "time"
    "github.com/e2u/e2util/e2time"
)

t := e2time.MustParse("2006-01-02", "2024-05-01")
next := e2time.AddDay(e2time.ToDay(), 1)
p := e2time.TimePointer(t)
```
