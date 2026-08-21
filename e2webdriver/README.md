# e2webdriver

下載並安裝對應作業系統的 ChromeDriver（Chrome for Testing）。

Download and install the platform ChromeDriver from Chrome for Testing.

不是 Selenium 會話封裝；只負責取得可執行檔路徑。

This is not a Selenium session wrapper; it only installs the binary.

下載測試需設定 `E2UTIL_WEBDRIVER_TEST=1`。

Download tests require `E2UTIL_WEBDRIVER_TEST=1`.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2webdriver
```

## 用法 / Usage

```go
import (
    "context"
    "github.com/e2u/e2util/e2webdriver"
)

path, err := e2webdriver.Install(context.Background(), "/usr/local/bin")
```
