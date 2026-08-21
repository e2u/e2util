# e2s3v2

遺留相容包，只保留 `ParseS3Path`。新代碼請用 [`e2aws/e2s3`](../../e2aws/e2s3)。

Leftover compatibility package with `ParseS3Path` only. New code should use [`e2aws/e2s3`](../../e2aws/e2s3).

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2awsv2/e2s3v2"

s := &e2s3v2.S3{}
bucket, key, err := s.ParseS3Path("s3://bucket/path/file")
```
