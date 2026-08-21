# e2md5

MD5 hex，以及大檔案「頭 128 + 尾 128 字節」摘要。

MD5 hex, plus a head-128/tail-128 digest for large blobs.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2hash/e2md5
```

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2hash/e2md5"

hex := e2md5.MD5HexString([]byte("hello"))
ht := e2md5.HeadTailHex(largeFileBytes) // ≤1024 字節則整段 MD5
```
