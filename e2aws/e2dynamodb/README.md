# e2dynamodb

AWS SDK Go v2 的 DynamoDB 封裝：Put/Get/Delete、Scan/Query 分頁。

DynamoDB helpers on AWS SDK for Go v2: put/get/delete and paginated scan/query.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2aws/e2dynamodb
```

## 功能 / Features

- **鍵 / Keys**：`BuildKeyValue`、字串集 `StringSet`
- **單筆 / Item**：`Put`、`GetByPK`、`GetByPKAndSK`、`DeleteByPKAndSK`
- **分頁 / Pages**：`ScanPages`、`QueryPages`

## 用法 / Usage

```go
import (
    "github.com/e2u/e2util/e2aws"
    "github.com/e2u/e2util/e2aws/e2dynamodb"
)

cfg := e2aws.NewSession("us-east-1")
db := e2dynamodb.New("users", cfg)
err := db.Put(map[string]any{"pk": "u1", "name": "ada"})
```
