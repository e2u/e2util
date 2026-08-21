# e2awsv2

遺留 stub。AWS SDK v2 輔助已搬到 [`e2aws`](../e2aws)。

Leftover stub. AWS SDK v2 helpers now live in [`e2aws`](../e2aws).

`e2awsv2/e2s3v2` 仍提供 `ParseS3Path` 以相容舊代碼。

`e2awsv2/e2s3v2` still exposes `ParseS3Path` for compatibility.

## 請改用 / Prefer

```go
import "github.com/e2u/e2util/e2aws"
import "github.com/e2u/e2util/e2aws/e2s3"
```
