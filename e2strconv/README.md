# e2strconv

安全／Must 風格的字串轉數字、布林、Unix 時間。

Must-style parsing of ints, floats, bools, and Unix timestamps.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2strconv
```

## 功能 / Features

- **泛型整數 / Generic int**：`ParseInt[T](s)`
- **Must 系列**：`MustParseInt`、`MustParseInt64`、`MustParseInt16`、`MustParseUint`、`MustParseFloat`、`MustParseBool`
- **時間 / Time**：`MustParseStringUnixTime`

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2strconv"

n, err := e2strconv.ParseInt[int64](" 42 ")
f := e2strconv.MustParseFloat("3.14")
t := e2strconv.MustParseStringUnixTime("1710000000")
```
