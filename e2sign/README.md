# e2sign

把 OS 訊號對應到回呼，在背景 goroutine 裡處理。

Map OS signals to callbacks and handle them in a background goroutine.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2sign
```

## 用法 / Usage

```go
import (
    "os"
    "syscall"
    "github.com/e2u/e2util/e2sign"
)

e2sign.RegisterSignTask(map[os.Signal]func(){
    syscall.SIGTERM: func() { shutdown() },
    syscall.SIGINT:  func() { shutdown() },
})
```
