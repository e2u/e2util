# e2hash

用任意 `hash.Hash` 算 hex；可用 Reader 串流。MD5 專用見 [`e2md5`](e2md5)。

Hex digest from any `hash.Hash`, including streaming from a Reader. See [`e2md5`](e2md5) for MD5 helpers.

密碼雜湊在 `e2crypto`，不在本套件。

Password hashing lives in `e2crypto`, not here.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2hash
```

## 用法 / Usage

```go
import (
    "crypto/sha256"
    "github.com/e2u/e2util/e2hash"
)

hex := e2hash.HashHex([]byte("hello"), sha256.New)
hex, err := e2hash.HashHexReader(reader, sha256.New)
```
