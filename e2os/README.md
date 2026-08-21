# e2os

OS helpers: file existence, working directory, retries, process signals, systemd unit text, and host IPv4.

```bash
go get github.com/e2u/e2util/e2os
```

## File and working directory

```go
import "github.com/e2u/e2util/e2os"

if e2os.FileExists("config.toml") {
    // ...
}

_ = e2os.ChdirToAppRoot() // go test / go run → module root; binary → executable dir
dir, _ := e2os.GetExecDir()
wd := e2os.MustGetwd()
_ = e2os.ChangeWorkdir("/var/app")
```

Environment variables live in `e2env`, not this package.

## Retry, IP, systemd, signals

```go
err := e2os.RetryRun(3, time.Second, func(i int) error {
    return ping()
})

ip, err := e2os.ExternalIP()

unit, err := e2os.InitSystemdService()

err = e2os.SendSignalToProcess("my-daemon", syscall.SIGTERM)
```
