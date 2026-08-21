# e2os

作業系統輔助：檔案是否存在、工作目錄、重試、行程訊號、systemd unit、本機 IPv4。

OS helpers: file existence, working directory, retries, process signals, systemd unit text, host IPv4.

環境變數請用 `e2env`，不在本套件。

Environment variables live in `e2env`, not this package.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2os
```

## 功能 / Features

- **檔案／目錄 / Files**：`FileExists`、`ChdirToAppRoot`、`GetExecDir`、`MustGetwd`、`ChangeWorkdir`
- **重試 / Retry**：`RetryRun`（最後一次失敗不再 sleep）
- **網路／服務 / Net & service**：`ExternalIP`、`InitSystemdService`、`SendSignalToProcess`

## 用法 / Usage

```go
import (
    "syscall"
    "time"
    "github.com/e2u/e2util/e2os"
)

if e2os.FileExists("config.toml") {
    _ = e2os.ChdirToAppRoot()
}
err := e2os.RetryRun(3, time.Second, func(i int) error { return ping() })
ip, err := e2os.ExternalIP()
_ = e2os.SendSignalToProcess("my-daemon", syscall.SIGTERM)
```
