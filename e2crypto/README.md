# e2crypto Documentation

## 項目概覽 / Project Overview

`e2crypto` 是 `e2util` 工具庫中的一個子包，提供了一組用於加密和隨機數據生成的工具函數。它包含兩大功能模塊：
1. **RSA 密鑰管理**：支持生成 RSA 密鑰對、從 PEM 格式字符串解析公鑰和私鑰，以及將公鑰和私鑰導出為 PEM 格式字符串，適用於數據加密、數字簽名或身份驗證等場景。
2. **隨機數據生成**：支持生成隨機字符串、字節、數字、浮點數以及從切片中隨機選擇元素，適用於令牌生成、隨機抽樣或加密密鑰生成等場景。
此包基於 Go 的 `crypto/rand`、`crypto/rsa` 和 `crypto/x509` 包，確保安全性和可靠性。

`e2crypto` is a sub-package of the `e2util` library, providing a set of utility functions for cryptography and random data generation. It includes two main functional modules:
1. **RSA Key Management**: Supports generating RSA key pairs, parsing public and private keys from PEM format strings, and exporting public and private keys as PEM format strings, suitable for scenarios like data encryption, digital signatures, or authentication.
2. **Random Data Generation**: Supports generating random strings, bytes, numbers, floats, and selecting random elements from slices, suitable for token creation, random sampling, or cryptographic key generation.
This package is built on Go's `crypto/rand`, `crypto/rsa`, and `crypto/x509` packages, ensuring security and reliability.

---

## 使用方法 / Usage

### 1. 生成 RSA 密鑰對 / Generating an RSA Key Pair

使用 `GenerateRsaKeyPair` 生成一對 RSA 密鑰（私鑰和公鑰）。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
// 生成 RSA 密鑰對
privateKey, publicKey := e2crypto.GenerateRsaKeyPair()
fmt.Println("私鑰:", privateKey)
fmt.Println("公鑰:", publicKey)
}
```

Use `GenerateRsaKeyPair` to generate a pair of RSA keys (private and public).

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
// Generate an RSA key pair
privateKey, publicKey := e2crypto.GenerateRsaKeyPair()
fmt.Println("Private Key:", privateKey)
fmt.Println("Public Key:", publicKey)
}
```

### 2. 導出 RSA 私鑰為 PEM 格式 / Exporting RSA Private Key as PEM String

使用 `ExportRsaPrivateKeyAsPemStr` 將 RSA 私鑰導出為 PEM 格式字符串。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
privateKey, _ := e2crypto.GenerateRsaKeyPair()
// 導出私鑰為 PEM 格式
pemStr := e2crypto.ExportRsaPrivateKeyAsPemStr(privateKey)
fmt.Println("私鑰 PEM:", pemStr)
}
```

Use `ExportRsaPrivateKeyAsPemStr` to export an RSA private key as a PEM format string.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
privateKey, _ := e2crypto.GenerateRsaKeyPair()
// Export the private key as PEM format
pemStr := e2crypto.ExportRsaPrivateKeyAsPemStr(privateKey)
fmt.Println("Private Key PEM:", pemStr)
}
```

### 3. 導出 RSA 公鑰為 PEM 格式 / Exporting RSA Public Key as PEM String

使用 `ExportRsaPublicKeyAsPemStr` 將 RSA 公鑰導出為 PEM 格式字符串。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
_, publicKey := e2crypto.GenerateRsaKeyPair()
// 導出公鑰為 PEM 格式
pemStr, err := e2crypto.ExportRsaPublicKeyAsPemStr(publicKey)
if err != nil {
fmt.Println("導出公鑰失敗:", err)
return
}
fmt.Println("公鑰 PEM:", pemStr)
}
```

Use `ExportRsaPublicKeyAsPemStr` to export an RSA public key as a PEM format string.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
_, publicKey := e2crypto.GenerateRsaKeyPair()
// Export the public key as PEM format
pemStr, err := e2crypto.ExportRsaPublicKeyAsPemStr(publicKey)
if err != nil {
fmt.Println("Failed to export public key:", err)
return
}
fmt.Println("Public Key PEM:", pemStr)
}
```

### 4. 從 PEM 字符串解析 RSA 私鑰 / Parsing RSA Private Key from PEM String

使用 `ParseRsaPrivateKeyFromPemString` 從 PEM 格式字符串解析 RSA 私鑰。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
privateKey, _ := e2crypto.GenerateRsaKeyPair()
pemStr := e2crypto.ExportRsaPrivateKeyAsPemStr(privateKey)
// 從 PEM 字符串解析私鑰
parsedKey, err := e2crypto.ParseRsaPrivateKeyFromPemString(pemStr)
if err != nil {
fmt.Println("解析私鑰失敗:", err)
return
}
fmt.Println("解析出的私鑰:", parsedKey)
}
```

Use `ParseRsaPrivateKeyFromPemString` to parse an RSA private key from a PEM format string.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
privateKey, _ := e2crypto.GenerateRsaKeyPair()
pemStr := e2crypto.ExportRsaPrivateKeyAsPemStr(privateKey)
// Parse the private key from PEM string
parsedKey, err := e2crypto.ParseRsaPrivateKeyFromPemString(pemStr)
if err != nil {
fmt.Println("Failed to parse private key:", err)
return
}
fmt.Println("Parsed Private Key:", parsedKey)
}
```

### 5. 從 PEM 字符串解析 RSA 公鑰 / Parsing RSA Public Key from PEM String

使用 `ParseRsaPublicKeyFromPemStr` 從 PEM 格式字符串解析 RSA 公鑰。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
_, publicKey := e2crypto.GenerateRsaKeyPair()
pemStr, _ := e2crypto.ExportRsaPublicKeyAsPemStr(publicKey)
// 從 PEM 字符串解析公鑰
parsedKey, err := e2crypto.ParseRsaPublicKeyFromPemStr(pemStr)
if err != nil {
fmt.Println("解析公鑰失敗:", err)
return
}
fmt.Println("解析出的公鑰:", parsedKey)
}
```

Use `ParseRsaPublicKeyFromPemStr` to parse an RSA public key from a PEM format string.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
_, publicKey := e2crypto.GenerateRsaKeyPair()
pemStr, _ := e2crypto.ExportRsaPublicKeyAsPemStr(publicKey)
// Parse the public key from PEM string
parsedKey, err := e2crypto.ParseRsaPublicKeyFromPemStr(pemStr)
if err != nil {
fmt.Println("Failed to parse public key:", err)
return
}
fmt.Println("Parsed Public Key:", parsedKey)
}
```

### 6. 生成隨機字符串 / Generating a Random String

使用 `RandomString` 生成指定長度的隨機字符串，採用預定義的字符集。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
// 生成長度為 10 的隨機字符串
str, err := e2crypto.RandomString(10)
if err != nil {
fmt.Println("生成字符串失敗:", err)
return
}
fmt.Println("隨機字符串:", str) // 例如 "Kb9pL2mX7q"
}
```

Use `RandomString` to generate a random string of specified length using a predefined character set.

```go
package main

import (
" fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
// Generate a random string of length 10
str, err := e2crypto.RandomString(10)
if err != nil {
fmt.Println("Failed to generate string:", err)
return
}
fmt.Println("Random string:", str) // e.g., "Kb9pL2mX7q"
}
```

### 7. 生成隨機字節 / Generating Random Bytes

使用 `RandomBytes` 生成指定長度的隨機字節切片。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
// 生成 16 個隨機字節
bytes, err := e2crypto.RandomBytes(16)
if err != nil {
fmt.Println("生成字節失敗:", err)
return
}
fmt.Printf("隨機字節: %x\n", bytes) // 例如 "1a2b3c4d5e6f7g8h9i0j"
}
```

Use `RandomBytes` to generate a slice of random bytes of specified length.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
// Generate 16 random bytes
bytes, err := e2crypto.RandomBytes(16)
if err != nil {
fmt.Println("Failed to generate bytes:", err)
return
}
fmt.Printf("Random bytes: %x\n", bytes) // e.g., "1a2b3c4d5e6f7g8h9i0j"
}
```

### 8. 生成隨機數字 / Generating a Random Number

使用 `RandomNumber` 生成指定範圍內的隨機整數。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
// 生成 1 到 100 之間的隨機整數
num, err := e2crypto.RandomNumber(1, 100)
if err != nil {
fmt.Println("生成數字失敗:", err)
return
}
fmt.Println("隨機數字:", num) // 例如 42
}
```

Use `RandomNumber` to generate a random integer within a specified range.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
// Generate a random integer between 1 and 100
num, err := e2crypto.RandomNumber(1, 100)
if err != nil {
fmt.Println("Failed to generate number:", err)
return
}
fmt.Println("Random number:", num) // e.g., 42
}
```

### 9. 生成隨機浮點數 / Generating a Random Float

使用 `RandomFloat` 生成指定範圍內的隨機浮點數。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
// 生成 0.0 到 1.0 之間的隨機浮點數
f, err := e2crypto.RandomFloat(0.0, 1.0)
if err != nil {
fmt.Println("生成浮點數失敗:", err)
return
}
fmt.Printf("隨機浮點數: %.6f\n", f) // 例如 0.732145
}
```

Use `RandomFloat` to generate a random floating-point number within a specified range.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
// Generate a random float between 0.0 and 1.0
f, err := e2crypto.RandomFloat(0.0, 1.0)
if err != nil {
fmt.Println("Failed to generate float:", err)
return
}
fmt.Printf("Random float: %.6f\n", f) // e.g., 0.732145
}
```

### 10. 隨機選擇元素 / Selecting a Random Element

使用 `RandomElement` 從切片中隨機選擇一個元素。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
// 從切片中隨機選擇一個元素
items := []string{"蘋果", "香蕉", "橙子"}
item, err := e2crypto.RandomElement(items)
if err != nil {
fmt.Println("選擇元素失敗:", err)
return
}
fmt.Println("隨機元素:", item) // 例如 "香蕉"
}
```

Use `RandomElement` to pick a random element from a slice.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2crypto"
)

func main() {
// Select a random element from a slice
items := []string{"apple", "banana", "orange"}
item, err := e2crypto.RandomElement(items)
if err != nil {
fmt.Println("Failed to select element:", err)
return
}
fmt.Println("Random element:", item) // e.g., "banana"
}
```

---

## 安裝步驟 / Installation Steps

1. **確保 Go 環境**
確認已安裝 Go（建議使用 1.16 或以上版本），並設置好 `GOPATH`。
2. **下載項目**
在終端運行以下命令：
```bash
go get -u github.com/e2u/e2util/e2crypto
```
3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2crypto`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。

1. **Ensure Go Environment**
Confirm Go (version 1.16 or higher recommended) is installed and `GOPATH` is set.
2. **Download the Package**
Run the following command in your terminal:
```bash
go get -u github.com/e2u/e2util/e2crypto
```
3. **Verify Installation**
Import `github.com/e2u/e2util/e2crypto` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **RSA 密鑰對生成**：`GenerateRsaKeyPair` 生成 4096 位 RSA 密鑰對（私鑰和公鑰）。
- **RSA 密鑰導出**：`ExportRsaPrivateKeyAsPemStr` 和 `ExportRsaPublicKeyAsPemStr` 將私鑰和公鑰導出為 PEM 格式字符串。
- **RSA 密鑰解析**：`ParseRsaPrivateKeyFromPemString` 和 `ParseRsaPublicKeyFromPemStr` 從 PEM 格式字符串解析私鑰和公鑰。
- **隨機字符串生成**：`RandomString` 使用 62 個字符的字母表（A-Z, a-z, 0-9）生成安全的隨機字符串。
- **隨機字節**：`RandomBytes` 提供安全的隨機字節切片，適用於加密或其他用途。
- **隨機數字**：`RandomNumber` 在指定範圍內生成整數，支持泛型整數類型。
- **隨機浮點數**：`RandomFloat` 在指定範圍內生成浮點數，具有精細精度。
- **隨機元素選擇**：`RandomElement` 從任意切片中隨機挑選元素，支持泛型類型。
- **錯誤處理**：所有函數均針對無效輸入（例如負長度、空切片、無效範圍或無效密鑰格式）返回錯誤。

- **RSA Key Pair Generation**: `GenerateRsaKeyPair` generates a 4096-bit RSA key pair (private and public keys).
- **RSA Key Export**: `ExportRsaPrivateKeyAsPemStr` and `ExportRsaPublicKeyAsPemStr` export private and public keys as PEM format strings.
- **RSA Key Parsing**: `ParseRsaPrivateKeyFromPemString` and `ParseRsaPublicKeyFromPemStr` parse private and public keys from PEM format strings.
- **Random String Generation**: `RandomString` generates secure random strings using a 62-character alphabet (A-Z, a-z, 0-9).
- **Random Bytes**: `RandomBytes` provides a secure random byte slice for cryptographic or general use.
- **Random Numbers**: `RandomNumber` generates integers within a specified range, supporting generic integer types.
- **Random Floats**: `RandomFloat` produces floating-point numbers within a range, with fine-grained precision.
- **Random Element Selection**: `RandomElement` picks a random item from any slice, supporting generic types.
- **Error Handling**: All functions return errors for invalid inputs (e.g., negative length, empty slice, invalid range, or invalid key format).

---
