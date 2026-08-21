# e2s3

AWS SDK Go v2 的 S3 封裝：上傳、下載、預簽名、列出、複製、刪除。

S3 helpers on AWS SDK for Go v2: upload, download, presign, list, copy, delete.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2aws/e2s3
```

## 功能 / Features

- **路徑 / Path**：`ParseS3Path("s3://bucket/key")`
- **寫入 / Write**：`PutContentObject`、`Upload`、`UploadWithFilePath`
- **讀取 / Read**：`GetObject`、`ListBucketFiles`
- **預簽名 / Presign**：`PreSignedGetObjectURL`、`PreSignedPutObjectURL`
- **管理 / Manage**：`DeleteObject`、`CopyObject`

## 用法 / Usage

```go
import (
    "github.com/e2u/e2util/e2aws"
    "github.com/e2u/e2util/e2aws/e2s3"
)

cfg := e2aws.NewSession("us-east-1")
s := e2s3.New(cfg)
bucket, key, _ := s.ParseS3Path("s3://my-bucket/dir/file.txt")
_ = s.PutContentObject(bucket, key, []byte("hello"))
```
