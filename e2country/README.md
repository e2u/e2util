# e2country Documentation

## 項目概覽 / Project Overview

`e2country` 是 `e2util` 工具庫中的一個子包，提供了一個全球國家代碼數據集。它包含各國的電話區號（`CountryCode`）、ISO 3166-1 兩位代碼（`ISOCode2`）和三位代碼（`ISOCode3`），並以結構化形式存儲。此包適用於需要國家代碼信息的應用場景，例如國際電話號碼格式化、地理位置識別或數據標準化。

`e2country` is a sub-package of the `e2util` library, providing a dataset of global country codes. It includes phone country codes (`CountryCode`), ISO 3166-1 two-letter codes (`ISOCode2`), and three-letter codes (`ISOCode3`), stored in a structured format. This package is suitable for applications requiring country code information, such as international phone number formatting, geolocation identification, or data standardization.

---

## 使用方法 / Usage

### 1. 訪問國家代碼數據 / Accessing Country Code Data

Access the `CountryCodes` slice to retrieve country code information.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2country"
)

func main() {
// Iterate through the CountryCodes slice
for _, country := range e2country.CountryCodes {
fmt.Printf("Country: %s, Code: %s, ISO2: %s, ISO3: %s\n",
country.Country, country.CountryCode, country.ISOCode2, country.ISOCode3)
}
}
```

### 2. 查找特定國家代碼 / Finding a Specific Country Code

Search for a country by name to retrieve its codes.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2country"
)

func main() {
// Find a country by name
countryName := "Canada"
for _, country := range e2country.CountryCodes {
if country.Country == countryName {
fmt.Printf("Found: %s, Code: %s, ISO2: %s, ISO3: %s\n",
country.Country, country.CountryCode, country.ISOCode2, country.ISOCode3)
break
}
}
}
```

### 3. 根據電話區號查找國家 / Finding Countries by Phone Code

Retrieve all countries with a specific phone country code.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2country"
)

func main() {
// Find countries by phone country code
phoneCode := "1"
for _, country := range e2country.CountryCodes {
if country.CountryCode == phoneCode {
fmt.Printf("Country: %s, Code: %s, ISO2: %s, ISO3: %s\n",
country.Country, country.CountryCode, country.ISOCode2, country.ISOCode3)
}
}
}
```

### 4. 根據 ISO 代碼查找國家 / Finding a Country by ISO Code

Search for a country using its ISO 3166-1 two-letter or three-letter code.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2country"
)

func main() {
// Find a country by ISO2 code
isoCode2 := "US"
for _, country := range e2country.CountryCodes {
if country.ISOCode2 == isoCode2 {
fmt.Printf("Found: %s, Code: %s, ISO2: %s, ISO3: %s\n",
country.Country, country.CountryCode, country.ISOCode2, country.ISOCode3)
break
}
}
}
```

---

## 安裝步驟 / Installation Steps

1. **確保 Go 環境**
確認已安裝 Go（建議使用 1.16 或以上版本），並設置好 `GOPATH`。
2. **下載項目**
在終端運行以下命令：
```bash
go get -u github.com/e2u/e2util/e2country
```
3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2country`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。

1. **Ensure Go Environment**
Confirm Go (version 1.16 or higher recommended) is installed and `GOPATH` is set.
2. **Download the Package**
Run the following command in your terminal:
```bash
go get -u github.com/e2u/e2util/e2country
```
3. **Verify Installation**
Import `github.com/e2u/e2util/e2country` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **全球國家代碼數據**：`CountryCodes` 包含各國的電話區號、ISO 3166-1 兩位和三位代碼。
- **結構化存儲**：使用 `CountryCode` 結構存儲國家信息，便於訪問和查詢。
- **多維查詢**：支持通過國家名稱、電話區號或 ISO 代碼查找國家信息。
- **輕量級設計**：數據以靜態切片形式存儲，無需外部依賴，適合快速集成。

- **Global Country Code Data**: `CountryCodes` includes phone country codes, ISO 3166-1 two-letter, and three-letter codes for each country.
- **Structured Storage**: Uses the `CountryCode` struct to store country information, facilitating access and querying.
- **Multi-Dimensional Querying**: Supports searching by country name, phone code, or ISO code.
- **Lightweight Design**: Data is stored as a static slice, requiring no external dependencies, ideal for quick integration.

---
```
