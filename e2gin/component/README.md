---

# component Documentation

## 項目概覽 / Project Overview

`component` 是 `e2gin` 子包的一部分，位於 `github.com/e2u/e2util/e2gin/component`，提供了一組用於 Web 應用中分頁處理的工具函數。它主要功能模塊為：
- **分頁查詢**：支持從數據庫中查詢分頁數據，並生成 HTML 分頁欄，適用於基於 `gin` 和 `gorm` 的 Web 應用場景。
此包依賴 `github.com/gin-gonic/gin`、`gorm.io/gorm` 等庫，並整合了 `e2util/e2strconv` 等工具，提供高效的分頁解決方案。

`component` is a sub-package of `e2gin`, located at `github.com/e2u/e2util/e2gin/component`, offering tools for pagination in web applications. Its primary functional module is:
- **Pagination Query**: Supports querying paginated data from a database and generating an HTML pagination bar, suitable for web applications using `gin` and `gorm`.
This package depends on libraries like `github.com/gin-gonic/gin` and `gorm.io/gorm`, integrating tools such as `e2util/e2strconv` for efficient pagination solutions.

---

## 使用方法 / Usage

### 1. 查詢分頁數據並生成分頁欄 / Query Paginated Data and Generate Pagination Bar

```go
package main

import (
"github.com/e2u/e2util/e2gin/component"
"github.com/gin-gonic/gin"
"gorm.io/gorm"
)

func main() {
// 假設有一個 gin 上下文和 gorm 數據庫連接 / Assume a gin context and gorm database connection
c := &gin.Context{}
db := &gorm.DB{}

// 查詢分頁數據 / Query paginated data
type Photo struct{ ID int }
prs, err := component.PaginationList(c, &Photo{}, db, &component.PaginationOption{PrePage: 10})
if err != nil {
// 處理錯誤 / Handle error
return
}
// 使用分頁結果 / Use pagination result
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
go get -u github.com/e2u/e2util/e2gin/component
```

3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2gin/component`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。
Import `github.com/e2u/e2util/e2gin/component` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **分頁查詢**：`PaginationList` 從數據庫查詢分頁數據，支持自定義每頁數量、排序字段和方向，並返回結構化結果。
**Pagination Query**: `PaginationList` queries paginated data from a database, supporting custom page size, sort field, and direction, returning a structured result.

- **HTML 分頁欄**：自動生成帶鏈接的分頁 HTML，根據當前頁面和總頁數動態調整顯示範圍。
**HTML Pagination Bar**: Automatically generates a linked pagination HTML bar, dynamically adjusting the display range based on the current page and total pages.

- **查詢參數解析**：從 `gin.Context` 中解析查詢參數（如頁碼、每頁數量），提供靈活的配置選項。
**Query Parameter Parsing**: Parses query parameters (e.g., page number, page size) from `gin.Context`, offering flexible configuration options.

- **錯誤處理**：針對數據庫查詢失敗或無效參數返回錯誤，確保健壯性。
**Error Handling**: Returns errors for database query failures or invalid parameters, ensuring robustness.

---
