---

# resp Documentation

## 項目概覽 / Project Overview

`resp` 是 `e2gin` 子包的一部分，位於 `github.com/e2u/e2util/e2gin/resp`，提供用於標準化 HTTP 響應的工具函數，主要支持：
- **響應處理**：生成標準化的成功和錯誤響應，支持 `gin` 和 `react-admin` 格式。
此包依賴 `github.com/gin-gonic/gin` 和 `github.com/sirupsen/logrus`，提供一致的響應結構。

`resp` is a sub-package of `e2gin`, located at `github.com/e2u/e2util/e2gin/resp`, offering tools for standardizing HTTP responses, primarily supporting:
- **Response Handling**: Generates standardized success and error responses, compatible with `gin` and `react-admin` formats.
This package depends on `github.com/gin-gonic/gin` and `github.com/sirupsen/logrus`, providing a consistent response structure.

---

## 使用方法 / Usage

### 1. 返回成功響應 / Return Success Response

```go
package main

import (
"github.com/e2u/e2util/e2gin/resp"
"github.com/gin-gonic/gin"
)

func main() {
// 初始化 gin / Initialize gin
c := &gin.Context{}

// 返回成功響應 / Return success response
resp.SuccessWithJSON(c, resp.Success, gin.H{"key": "value"})
}
```

### 2. 返回錯誤響應 / Return Error Response

```go
package main

import (
"github.com/e2u/e2util/e2gin/resp"
"github.com/gin-gonic/gin"
)

func main() {
// 初始化 gin / Initialize gin
c := &gin.Context{}

// 返回錯誤響應 / Return error response
resp.AboutWithJSON(c, resp.NotFound, "資源未找到 / Resource not found")
}
```

---

## 安裝步驟 / Installation Steps

1. **確保 Go 環境**
確認已安裝 Go（建議使用 1.16 或以上版本），並設置好 `GOPATH`。
Ensure Go (version 1.16 or higher recommended) is installed and `GOPATH` is set.

2. **下載項目**
在終端運行以下命令 / Run the following command in your terminal:
```bash
go get -u github.com/e2u/e2util/e2gin/resp
```

3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2gin/resp`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。
Import `github.com/e2u/e2util/e2gin/resp` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **成功響應**：`SuccessWithJSON` 生成標準化成功響應，支持自定義數據和 `react-admin` 格式。
**Success Response**: `SuccessWithJSON` generates standardized success responses, supporting custom data and `react-admin` format.

- **錯誤響應**：`AboutWithJSON` 生成標準化錯誤響應，支持詳細錯誤信息。
**Error Response**: `AboutWithJSON` generates standardized error responses, supporting detailed error information.

- **格式適配**：自動檢測 `X-Api-Consumer` 頭，支持 `react-admin` 的數據格式並設置 `X-Total-Count`。
**Format Adaptation**: Automatically detects the `X-Api-Consumer` header, supporting `react-admin` data format and setting `X-Total-Count`.

- **錯誤處理**：針對無效數據類型或反射異常進行恢復，確保響應穩定性。
**Error Handling**: Recovers from invalid data types or reflection errors, ensuring response stability.

---
