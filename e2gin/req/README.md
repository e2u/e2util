---

# req Documentation

## 項目概覽 / Project Overview

`req` 是 `e2gin` 子包的一部分，位於 `github.com/e2u/e2util/e2gin/req`，提供用於解析 RESTful API 請求參數的工具函數，主要支持：
- **參數解析**：解析排序、分頁和過濾參數，適用於基於 REST API 的數據查詢場景。
此包依賴 `github.com/e2u/e2util/e2json` 和 `golang.org/x/exp/maps`，提供高效的參數處理能力。

`req` is a sub-package of `e2gin`, located at `github.com/e2u/e2util/e2gin/req`, offering tools for parsing RESTful API request parameters, primarily supporting:
- **Parameter Parsing**: Parses sort, pagination, and filter parameters, suitable for REST API data query scenarios.
This package depends on `github.com/e2u/e2util/e2json` and `golang.org/x/exp/maps`, providing efficient parameter handling.

---

## 使用方法 / Usage

### 1. 解析排序參數 / Parse Sort Parameters

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2gin/req"
)

func main() {
// 解析排序參數 / Parse sort parameters
sort, _ := req.ParseSortPayload(`["published_at", "DESC"]`)
fmt.Println("字段 / Field:", sort.Field, "順序 / Order:", sort.Order)
}
```

### 2. 解析分頁參數 / Parse Pagination Parameters

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2gin/req"
)

func main() {
// 解析分頁參數 / Parse pagination parameters
pagination, _ := req.ParsePaginationPayload(`{"page": 2, "perPage": 20}`)
fmt.Println("頁碼 / Page:", pagination.Page, "每頁數量 / PerPage:", pagination.PrePage)
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
go get -u github.com/e2u/e2util/e2gin/req
```

3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2gin/req`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。
Import `github.com/e2u/e2util/e2gin/req` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **排序解析**：`ParseSortPayload` 解析排序字段和方向，支持 JSON 格式輸入。
**Sort Parsing**: `ParseSortPayload` parses sort field and direction, supporting JSON format input.

- **分頁解析**：`ParsePaginationPayload` 解析頁碼和每頁數量，支持默認值。
**Pagination Parsing**: `ParsePaginationPayload` parses page number and per-page count, supporting defaults.

- **過濾解析**：`ParseFilterPayload` 解析過濾條件，支持多種操作符（如等於、大於）。
**Filter Parsing**: `ParseFilterPayload` parses filter conditions, supporting multiple operators (e.g., equals, greater than).

- **範圍解析**：`ParseRangePayload` 解析範圍參數，支持正則和 JSON 格式。
**Range Parsing**: `ParseRangePayload` parses range parameters, supporting regex and JSON formats.

---
