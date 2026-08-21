# e2run

在 goroutine 裡跑一次或循環跑函式（可帶 context、sleep）。

Run a function once or in a loop on a goroutine, with optional context and sleep.

子目錄 `groupedcmd` 目前為空佔位。

The `groupedcmd` subpackage is currently an empty stub.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2run
```

## 功能 / Features

- `GoRunner`：背景跑一次
- `GoLoopRunner`：固定秒數間隔循環
- `GoLoopRunnerContext`：`ctx.Done()` 時停止
- `GoLoopRunnerWithoutSleep`：忙等循環

## 用法 / Usage

```go
import (
    "context"
    "github.com/e2u/e2util/e2run"
)

e2run.GoRunner(func() { doWork() })
ctx, cancel := context.WithCancel(context.Background())
e2run.GoLoopRunnerContext(ctx, func() { poll() }, 5)
defer cancel()
```
