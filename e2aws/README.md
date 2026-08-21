# e2aws

AWS SDK Go v2 共用設定，以及本機主機名／IPv4。S3、SQS、DynamoDB 見子套件。

Shared AWS SDK for Go v2 config, plus hostname/IPv4. See subpackages for S3, SQS, DynamoDB.

舊的 `e2awsv2` 是遺留 stub，新代碼請用本套件。

The old `e2awsv2` package is a leftover stub; new code should use `e2aws`.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2aws
```

## 功能 / Features

- **連線 / Session**：`NewSession(region, optFns...)` → `aws.Config`
- **主機 / Host**：`GetHostName`、`MustGetHostName`、`GetIP`、`MustGetIP`
- **子套件 / Subpackages**：[`e2s3`](e2s3)、[`e2sqs`](e2sqs)、[`e2dynamodb`](e2dynamodb)
- **UploadToS3**：仍可用，但建議改走 `e2s3`

## 用法 / Usage

```go
import (
    "github.com/e2u/e2util/e2aws"
    "github.com/e2u/e2util/e2aws/e2s3"
)

cfg := e2aws.NewSession("us-east-1")
s := e2s3.New(cfg)
_ = s.PutContentObject("bucket", "key.txt", []byte("hi"))
```
