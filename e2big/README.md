# e2big Documentation

## 項目概覽 / Project Overview

`e2big` 是 `e2util` 工具庫中的一個子包，封裝了大數計算功能，基於 `shopspring/decimal` 庫實現。它提供了浮點數的高精度加、減、乘、除和取模運算，支持兩個或多個數字的連續運算。此包適用於需要高精度浮點數計算的場景，例如金融應用、科學計算或其他對精度要求較高的應用。

`e2big` is a sub-package of the `e2util` library, encapsulating big number arithmetic operations based on the `shopspring/decimal` library. It provides high-precision addition, subtraction, multiplication, division, and modulus operations for floating-point numbers, supporting sequential operations on two or more numbers. This package is suitable for scenarios requiring high-precision floating-point calculations, such as financial applications, scientific computations, or other precision-sensitive applications.

---

## 使用方法 / Usage

### 1. 浮點大數加法 / Floating-Point Big Number Addition

使用 `FloatAdd` 進行兩個浮點數的高精度加法運算。

Use `FloatAdd` to perform high-precision addition on two floating-point numbers.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2big"
)

func main() {
// Add two floating-point numbers
result := e2big.FloatAdd(123.456, 789.123)
fmt.Println("Addition result:", result.String()) // e.g., "912.579"
}
```

### 2. 多個浮點數連續加法 / Sequential Addition of Multiple Floating-Point Numbers

使用 `FloatAddN` 對多個浮點數從左到右進行連續加法運算。

Use `FloatAddN` to perform sequential addition on multiple floating-point numbers from left to right.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2big"
)

func main() {
// Sequentially add multiple floating-point numbers
result := e2big.FloatAddN(123.456, 789.123, 456.789)
fmt.Println("Sequential addition result:", result.String()) // e.g., "1369.368"
}
```

### 3. 浮點大數減法 / Floating-Point Big Number Subtraction

使用 `FloatSub` 進行兩個浮點數的高精度減法運算。

Use `FloatSub` to perform high-precision subtraction on two floating-point numbers.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2big"
)

func main() {
// Subtract two floating-point numbers
result := e2big.FloatSub(789.123, 123.456)
fmt.Println("Subtraction result:", result.String()) // e.g., "665.667"
}
```

### 4. 多個浮點數連續減法 / Sequential Subtraction of Multiple Floating-Point Numbers

使用 `FloatSubN` 對多個浮點數從左到右進行連續減法運算。

Use `FloatSubN` to perform sequential subtraction on multiple floating-point numbers from left to right.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2big"
)

func main() {
// Sequentially subtract multiple floating-point numbers
result := e2big.FloatSubN(1000.0, 123.456, 456.789)
fmt.Println("Sequential subtraction result:", result.String()) // e.g., "419.755"
}
```

### 5. 浮點大數乘法 / Floating-Point Big Number Multiplication

使用 `FloatMul` 進行兩個浮點數的高精度乘法運算。

Use `FloatMul` to perform high-precision multiplication on two floating-point numbers.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2big"
)

func main() {
// Multiply two floating-point numbers
result := e2big.FloatMul(123.456, 2.5)
fmt.Println("Multiplication result:", result.String()) // e.g., "308.64"
}
```

### 6. 多個浮點數連續乘法 / Sequential Multiplication of Multiple Floating-Point Numbers

使用 `FloatMulN` 對多個浮點數從左到右進行連續乘法運算。

Use `FloatMulN` to perform sequential multiplication on multiple floating-point numbers from left to right.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2big"
)

func main() {
// Sequentially multiply multiple floating-point numbers
result := e2big.FloatMulN(123.456, 2.5, 3.0)
fmt.Println("Sequential multiplication result:", result.String()) // e.g., "925.92"
}
```

### 7. 浮點大數除法 / Floating-Point Big Number Division

使用 `FloatDiv` 進行兩個浮點數的高精度除法運算。

Use `FloatDiv` to perform high-precision division on two floating-point numbers.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2big"
)

func main() {
// Divide two floating-point numbers
result := e2big.FloatDiv(123.456, 2.0)
fmt.Println("Division result:", result.String()) // e.g., "61.728"
}
```

### 8. 多個浮點數連續除法 / Sequential Division of Multiple Floating-Point Numbers

使用 `FloatDivN` 對多個浮點數從左到右進行連續除法運算。

Use `FloatDivN` to perform sequential division on multiple floating-point numbers from left to right.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2big"
)

func main() {
// Sequentially divide multiple floating-point numbers
result := e2big.FloatDivN(123.456, 2.0, 3.0)
fmt.Println("Sequential division result:", result.String()) // e.g., "20.576"
}
```

### 9. 浮點大數取模 / Floating-Point Big Number Modulus

使用 `FloatMod` 進行兩個浮點數的高精度取模運算。

Use `FloatMod` to perform high-precision modulus operation on two floating-point numbers.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2big"
)

func main() {
// Perform modulus on two floating-point numbers
result := e2big.FloatMod(123.456, 10.0)
fmt.Println("Modulus result:", result.String()) // e.g., "3.456"
}
```

---

## 安裝步驟 / Installation Steps

1. **確保 Go 環境**
確認已安裝 Go（建議使用 1.16 或以上版本），並設置好 `GOPATH`。
2. **下載項目**
在終端運行以下命令：
```bash
go get -u github.com/e2u/e2util/e2big
```
3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2big`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。

1. **Ensure Go Environment**
Confirm Go (version 1.16 or higher recommended) is installed and `GOPATH` is set.
2. **Download the Package**
Run the following command in your terminal:
```bash
go get -u github.com/e2u/e2util/e2big
```
3. **Verify Installation**
Import `github.com/e2u/e2util/e2big` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **高精度加法**：`FloatAdd` 和 `FloatAddN` 提供兩個或多個浮點數的高精度加法運算。
- **高精度減法**：`FloatSub` 和 `FloatSubN` 提供兩個或多個浮點數的高精度減法運算。
- **高精度乘法**：`FloatMul` 和 `FloatMulN` 提供兩個或多個浮點數的高精度乘法運算。
- **高精度除法**：`FloatDiv` 和 `FloatDivN` 提供兩個或多個浮點數的高精度除法運算。
- **高精度取模**：`FloatMod` 提供兩個浮點數的高精度取模運算。
- **零值常量**：`Zero` 提供一個高精度的零值常量，用於比較或初始化。

- **High-Precision Addition**: `FloatAdd` and `FloatAddN` provide high-precision addition for two or more floating-point numbers.
- **High-Precision Subtraction**: `FloatSub` and `FloatSubN` provide high-precision subtraction for two or more floating-point numbers.
- **High-Precision Multiplication**: `FloatMul` and `FloatMulN` provide high-precision multiplication for two or more floating-point numbers.
- **High-Precision Division**: `FloatDiv` and `FloatDivN` provide high-precision division for two or more floating-point numbers.
- **High-Precision Modulus**: `FloatMod` provides high-precision modulus operation for two floating-point numbers.
- **Zero Constant**: `Zero` provides a high-precision zero constant for comparison or initialization.

---
```
